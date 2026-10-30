package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/diamondburned/oggreader"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

const (
	// maxVoicePlayers is how many servers can have the bot connected for audio
	// at once, across all apps of this process. Each one keeps a voice
	// connection open and runs ffmpeg while it plays.
	maxVoicePlayers = 200
	// maxAudioDuration is how much of an audio file is played at most.
	maxAudioDuration = 15 * time.Minute
	// voiceJoinTimeout is how long joining a voice channel may take.
	voiceJoinTimeout = 15 * time.Second
	// voiceLeaveTimeout is how long leaving a voice channel may take.
	voiceLeaveTimeout = 10 * time.Second
	// silenceFrames is how many Opus silence frames are sent after the audio,
	// so clients don't interpolate the last frame.
	silenceFrames = 5
)

var errFFmpegMissing = errors.New("playing audio needs ffmpeg, which isn't installed on the server")

// opusSilence is an Opus frame of silence.
var opusSilence = []byte{0xF8, 0xFF, 0xFE}

// VoiceManager keeps the voice connections the bots use to play audio, one per
// app and server. Playback outlives the flow that started it.
type VoiceManager struct {
	mu      sync.Mutex
	players map[voiceKey]*voicePlayer
}

func NewVoiceManager() *VoiceManager {
	return &VoiceManager{
		players: make(map[voiceKey]*voicePlayer),
	}
}

type voiceKey struct {
	appID   string
	guildID discord.GuildID
}

// voicePlayer is the voice connection of a bot in a server.
type voicePlayer struct {
	key voiceKey

	// mu serializes playing, stopping and leaving in the server.
	mu sync.Mutex
	// removed is set once the player is no longer in the manager. Guarded by
	// VoiceManager.mu.
	removed bool

	session   *state.State
	voice     *voice.Session
	channelID discord.ChannelID
	connected bool

	// cancel stops the current playback, done is closed once it stopped.
	cancel context.CancelFunc
	done   chan struct{}
	// left is set by the playback before closing done, if it left the voice
	// channel.
	left bool
}

// acquire returns the locked player of the server, creating it if there's
// none.
func (m *VoiceManager) acquire(key voiceKey) (*voicePlayer, error) {
	for {
		m.mu.Lock()
		p, ok := m.players[key]
		if !ok {
			if len(m.players) >= maxVoicePlayers {
				m.mu.Unlock()
				return nil, fmt.Errorf("too many bots are playing audio right now, try again later")
			}
			p = &voicePlayer{key: key}
			m.players[key] = p
		}
		m.mu.Unlock()

		p.mu.Lock()

		m.mu.Lock()
		removed := p.removed
		m.mu.Unlock()
		if !removed {
			return p, nil
		}

		// The playback left the channel in the meantime, start over with a
		// new player.
		p.mu.Unlock()
	}
}

// lookup returns the locked player of the server, or nil if there's none.
func (m *VoiceManager) lookup(key voiceKey) *voicePlayer {
	m.mu.Lock()
	p := m.players[key]
	m.mu.Unlock()
	if p == nil {
		return nil
	}

	p.mu.Lock()

	m.mu.Lock()
	removed := p.removed
	m.mu.Unlock()
	if removed {
		p.mu.Unlock()
		return nil
	}

	return p
}

func (m *VoiceManager) remove(p *voicePlayer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p.removed = true
	if m.players[p.key] == p {
		delete(m.players, p.key)
	}
}

// Play joins the voice channel and starts playing the audio in the background.
func (m *VoiceManager) Play(ctx context.Context, appID string, session *state.State, opts provider.PlayAudioOpts) error {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return errFFmpegMissing
	}

	key := voiceKey{appID: appID, guildID: opts.GuildID}
	p, err := m.acquire(key)
	if err != nil {
		return err
	}
	defer p.mu.Unlock()

	p.stopPlayback()

	// Moving the voice session to another channel works, but reconnecting
	// from scratch is more reliable, and the session of a restarted app is
	// gone anyway.
	if p.connected && (p.session != session || p.channelID != opts.ChannelID) {
		p.leave()
	}

	if !p.connected {
		if err := p.join(ctx, session, opts.ChannelID); err != nil {
			m.remove(p)
			return err
		}
	}

	playCtx, cancel := context.WithTimeout(context.Background(), maxAudioDuration+time.Minute)
	done := make(chan struct{})
	p.cancel = cancel
	p.done = done

	logger := slog.With(
		slog.String("app_id", appID),
		slog.String("guild_id", opts.GuildID.String()),
	)

	v := p.voice
	go func() {
		defer close(done)
		defer cancel()

		err := streamAudio(playCtx, v, ffmpeg, opts)
		// Stopped by StopAudio, Leave or the next Play, which decide what
		// happens to the connection.
		if errors.Is(playCtx.Err(), context.Canceled) {
			return
		}
		if err != nil {
			logger.Warn("Failed to play audio", slog.String("error", err.Error()))
		}

		// A failed playback leaves as well, the connection may be broken.
		// The player's fields are reset by stopPlayback, which waits for done.
		if opts.LeaveAfter || err != nil {
			leaveVoice(v)
			p.left = true
			m.remove(p)
		}
	}()

	return nil
}

// Stop stops the audio the bot plays in the server.
func (m *VoiceManager) Stop(appID string, guildID discord.GuildID) {
	p := m.lookup(voiceKey{appID: appID, guildID: guildID})
	if p == nil {
		return
	}
	defer p.mu.Unlock()

	p.stopPlayback()
}

// Leave stops the audio and leaves the voice channel the bot joined to play
// audio in the server. It reports whether the bot was connected for audio.
func (m *VoiceManager) Leave(appID string, guildID discord.GuildID) bool {
	p := m.lookup(voiceKey{appID: appID, guildID: guildID})
	if p == nil {
		return false
	}
	defer p.mu.Unlock()

	p.stopPlayback()
	connected := p.connected
	if connected {
		p.leave()
	}
	m.remove(p)

	return connected
}

func (p *voicePlayer) join(ctx context.Context, session *state.State, channelID discord.ChannelID) error {
	v, err := voice.NewSession(session)
	if err != nil {
		return fmt.Errorf("failed to create voice session: %w", err)
	}
	v.DAVELogger = slog.With(
		slog.String("app_id", p.key.appID),
		slog.String("guild_id", p.key.guildID.String()),
	)

	ctx, cancel := context.WithTimeout(ctx, voiceJoinTimeout)
	defer cancel()

	// Deafened, the bot only talks.
	if err := v.JoinChannelAndSpeak(ctx, channelID, false, true); err != nil {
		leaveVoice(v)
		return fmt.Errorf("failed to join the voice channel: %w", err)
	}

	p.session = session
	p.voice = v
	p.channelID = channelID
	p.connected = true
	return nil
}

func (p *voicePlayer) leave() {
	leaveVoice(p.voice)
	p.voice = nil
	p.session = nil
	p.channelID = 0
	p.connected = false
}

func leaveVoice(v *voice.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), voiceLeaveTimeout)
	defer cancel()

	if err := v.Leave(ctx); err != nil {
		slog.Debug("Failed to leave voice channel", slog.String("error", err.Error()))
	}
}

// stopPlayback stops the current playback and waits until it stopped.
func (p *voicePlayer) stopPlayback() {
	if p.done == nil {
		return
	}

	p.cancel()
	<-p.done
	p.cancel = nil
	p.done = nil

	if p.left {
		p.left = false
		p.voice = nil
		p.session = nil
		p.channelID = 0
		p.connected = false
	}
}

// streamAudio converts the audio to Opus with ffmpeg and sends it to the voice
// connection until it ended or ctx is done.
func streamAudio(ctx context.Context, v io.Writer, ffmpeg string, opts provider.PlayAudioOpts) error {
	cmd := exec.CommandContext(ctx, ffmpeg,
		"-hide_banner", "-loglevel", "error",
		// Only read the input from stdin. Formats like playlists can
		// reference other files and URLs, which mustn't be opened.
		"-protocol_whitelist", "pipe",
		"-i", "pipe:0",
		"-t", strconv.Itoa(int(maxAudioDuration.Seconds())),
		"-map", "0:a:0",
		"-filter:a", "volume="+strconv.FormatFloat(opts.Volume, 'f', 2, 64),
		"-c:a", "libopus",
		"-b:a", "96k",
		"-ar", "48000",
		"-ac", "2",
		// Matches the 20ms frames the voice connection sends.
		"-frame_duration", "20",
		"-application", "audio",
		// Playback is real time, one thread keeps up easily.
		"-threads", "1",
		"-f", "opus",
		"pipe:1",
	)
	cmd.Stdin = bytes.NewReader(opts.Audio)

	var stderr limitedBuffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg output: %w", err)
	}

	if err := cmd.Start(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	decodeErr := oggreader.DecodeBuffered(opusFrameWriter{v}, stdout)
	if decodeErr != nil {
		// ffmpeg would block writing output nobody reads.
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()

	if ctx.Err() != nil {
		return nil
	}
	if waitErr != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("ffmpeg failed: %s", msg)
		}
		return fmt.Errorf("ffmpeg failed: %w", waitErr)
	}
	if decodeErr != nil {
		return fmt.Errorf("failed to send audio: %w", decodeErr)
	}

	for range silenceFrames {
		if _, err := v.Write(opusSilence); err != nil {
			return nil
		}
	}

	return nil
}

// opusFrameWriter writes the Opus frames of an Ogg stream to the voice
// connection, skipping the header packets.
type opusFrameWriter struct {
	w io.Writer
}

func (w opusFrameWriter) Write(b []byte) (int, error) {
	if len(b) == 0 || bytes.HasPrefix(b, []byte("OpusHead")) || bytes.HasPrefix(b, []byte("OpusTags")) {
		return len(b), nil
	}
	return w.w.Write(b)
}

// maxStderrSize is how much of ffmpeg's error output is kept.
const maxStderrSize = 1024

// limitedBuffer keeps the first maxStderrSize bytes written to it.
type limitedBuffer struct {
	buf bytes.Buffer
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if room := maxStderrSize - b.buf.Len(); room > 0 {
		b.buf.Write(p[:min(len(p), room)])
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}
