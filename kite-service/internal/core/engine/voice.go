package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/voice"
	davesession "github.com/thomas-vilte/dave-go/session"

	"github.com/kitecloud/kite/kite-service/internal/core/sound"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

const (
	// maxVoicePlayers is how many servers can have the bot connected for audio
	// at once, across all apps of this process. Each one keeps a voice
	// connection open.
	maxVoicePlayers = 200
	// voiceJoinTimeout is how long joining a voice channel may take.
	voiceJoinTimeout = 15 * time.Second
	// voiceLeaveTimeout is how long leaving a voice channel may take.
	voiceLeaveTimeout = 10 * time.Second
)

// VoiceManager keeps the voice connections the bots use to play audio, one per
// app and server. Playback outlives the flow that started it.
type VoiceManager struct {
	assets store.AssetStore

	mu      sync.Mutex
	players map[voiceKey]*voicePlayer
}

func NewVoiceManager(assets store.AssetStore) *VoiceManager {
	return &VoiceManager{
		assets:  assets,
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

// Play joins the voice channel and starts playing the sound in the background.
func (m *VoiceManager) Play(ctx context.Context, appID string, session *state.State, opts provider.PlayAudioOpts) error {
	asset, err := m.assets.AssetWithContent(ctx, opts.AssetID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && asset.AppID != appID) {
		return fmt.Errorf("the sound doesn't exist anymore, upload it again")
	}
	if err != nil {
		return fmt.Errorf("failed to get the sound: %w", err)
	}
	if asset.ContentType != sound.ContentType {
		return fmt.Errorf("the file isn't a sound, upload it again in the block")
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

	playCtx, cancel := context.WithTimeout(context.Background(), sound.MaxDuration+time.Minute)
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

		err := sound.Play(playCtx, v, asset.Content)
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
	v.DAVE = davesession.CreateFunc(davesession.WithLogger(slog.With(
		slog.String("app_id", p.key.appID),
		slog.String("guild_id", p.key.guildID.String()),
	)))

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
