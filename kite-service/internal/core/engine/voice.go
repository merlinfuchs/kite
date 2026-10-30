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
	maxVoicePlayers = 1000
	// voiceIdleTimeout is how long the bot stays in the voice channel after the
	// sound ended or was stopped, until it leaves by itself.
	voiceIdleTimeout = 15 * time.Minute
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

func (k voiceKey) logger() *slog.Logger {
	return slog.With(
		slog.String("app_id", k.appID),
		slog.String("guild_id", k.guildID.String()),
	)
}

// voicePlayer is the voice connection of a bot in a server.
type voicePlayer struct {
	key voiceKey

	// mu serializes playing, stopping and leaving in the server.
	mu sync.Mutex

	session *state.State
	// voice is nil while the bot isn't connected.
	voice     *voice.Session
	channelID discord.ChannelID

	// The playback plays the sound, waits while the bot is idle, then leaves
	// the channel and removes the player. cancel ends it and done is closed
	// once it ended. stopSound only stops the sound.
	cancel    context.CancelFunc
	stopSound context.CancelFunc
	done      chan struct{}
}

// player returns the locked player of the server. If there's none, it creates
// one if create is set, and returns nil otherwise.
func (m *VoiceManager) player(key voiceKey, create bool) (*voicePlayer, error) {
	for {
		m.mu.Lock()
		p, ok := m.players[key]
		if !ok {
			if !create {
				m.mu.Unlock()
				return nil, nil
			}
			if len(m.players) >= maxVoicePlayers {
				m.mu.Unlock()
				return nil, fmt.Errorf("too many bots are playing audio right now, try again later")
			}
			p = &voicePlayer{key: key}
			m.players[key] = p
		}
		m.mu.Unlock()

		p.mu.Lock()
		if m.has(p) {
			return p, nil
		}

		// The playback left the channel in the meantime, start over.
		p.mu.Unlock()
	}
}

func (m *VoiceManager) has(p *voicePlayer) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.players[p.key] == p
}

func (m *VoiceManager) remove(p *voicePlayer) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	var p *voicePlayer
	for {
		p, err = m.player(key, true)
		if err != nil {
			return err
		}

		p.stopPlayback()
		if m.has(p) {
			break
		}

		// The playback we stopped had already left the channel.
		p.mu.Unlock()
	}
	defer p.mu.Unlock()

	// Moving the voice session to another channel works, but reconnecting
	// from scratch is more reliable, and the session of a restarted app is
	// gone anyway.
	if p.voice != nil && (p.session != session || p.channelID != opts.ChannelID) {
		p.leave()
	}

	if p.voice == nil {
		if err := p.join(ctx, session, opts.ChannelID); err != nil {
			m.remove(p)
			return err
		}
	}

	playCtx, cancel := context.WithCancel(context.Background())
	soundCtx, stopSound := context.WithTimeout(playCtx, sound.MaxDuration+time.Minute)
	done := make(chan struct{})
	p.cancel = cancel
	p.stopSound = stopSound
	p.done = done

	// The sound is an argument instead of captured, so it can be freed while
	// the bot waits in the channel.
	go func(v *voice.Session, content []byte) {
		defer close(done)
		defer cancel()
		defer stopSound()

		err := sound.Play(soundCtx, v, content)
		// Ended by the next Play or by Leave, which decide what happens to the
		// connection.
		if playCtx.Err() != nil {
			return
		}
		if err != nil {
			key.logger().Warn("Failed to play audio", slog.String("error", err.Error()))
		}

		// Stop audio keeps the bot in the channel, like Leave When Done being
		// off. A failed playback leaves right away, the connection may be
		// broken.
		stopped := errors.Is(soundCtx.Err(), context.Canceled)
		if err == nil && (stopped || !opts.LeaveAfter) {
			select {
			case <-playCtx.Done():
				return
			case <-time.After(voiceIdleTimeout):
			}
		}

		// Nobody else touches the player until done is closed, see
		// stopPlayback.
		p.leave()
		m.remove(p)
	}(p.voice, asset.Content)

	return nil
}

// Stop stops the sound the bot plays in the server. The bot stays in the
// channel until it's idle for too long.
func (m *VoiceManager) Stop(appID string, guildID discord.GuildID) {
	p, _ := m.player(voiceKey{appID: appID, guildID: guildID}, false)
	if p == nil {
		return
	}
	defer p.mu.Unlock()

	if p.stopSound != nil {
		p.stopSound()
	}
}

// Leave stops the sound and closes the voice connection the bot joined to play
// audio in the server.
func (m *VoiceManager) Leave(appID string, guildID discord.GuildID) {
	p, _ := m.player(voiceKey{appID: appID, guildID: guildID}, false)
	if p == nil {
		return
	}
	defer p.mu.Unlock()

	p.stopPlayback()
	if p.voice != nil {
		p.leave()
	}
	m.remove(p)
}

func (p *voicePlayer) join(ctx context.Context, session *state.State, channelID discord.ChannelID) error {
	v, err := voice.NewSession(session)
	if err != nil {
		return fmt.Errorf("failed to create voice session: %w", err)
	}
	v.DAVE = davesession.CreateFunc(davesession.WithLogger(p.key.logger()))

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
	return nil
}

func (p *voicePlayer) leave() {
	leaveVoice(p.voice)
	p.voice = nil
}

func leaveVoice(v *voice.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), voiceLeaveTimeout)
	defer cancel()

	if err := v.Leave(ctx); err != nil {
		slog.Debug("Failed to leave voice channel", slog.String("error", err.Error()))
	}
}

// stopPlayback ends the playback and waits until it ended. Until then, the
// playback may leave the channel and change the player.
func (p *voicePlayer) stopPlayback() {
	if p.done == nil {
		return
	}

	p.cancel()
	<-p.done
	p.cancel = nil
	p.stopSound = nil
	p.done = nil
}
