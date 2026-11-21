package flow

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ttsTestSpeaker struct {
	calls []provider.SpeakOpts
	err   error
}

func (s *ttsTestSpeaker) Speak(ctx context.Context, opts provider.SpeakOpts) (provider.TTSAudio, error) {
	s.calls = append(s.calls, opts)
	if s.err != nil {
		return provider.TTSAudio{}, s.err
	}
	return provider.TTSAudio{Content: []byte("sound of " + opts.Text)}, nil
}

type ttsTestDiscord struct {
	provider.MockDiscordProvider

	channel *discord.Channel
	played  *provider.PlayAudioOpts
	playErr error
}

func (p *ttsTestDiscord) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	return p.channel, nil
}

func (p *ttsTestDiscord) PlayAudio(ctx context.Context, opts provider.PlayAudioOpts) error {
	if p.playErr != nil {
		return p.playErr
	}
	p.played = &opts
	return nil
}

func newTTSTestDiscord() *ttsTestDiscord {
	return &ttsTestDiscord{channel: &discord.Channel{ID: 3, GuildID: 5, Type: discord.GuildVoice}}
}

func executeTTSBlock(t *testing.T, d provider.DiscordProvider, tts provider.TTSProvider, dataJSON string) error {
	var data FlowNodeData
	require.NoError(t, json.Unmarshal([]byte(dataJSON), &data))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&blockTestContextData{},
		FlowProviders{
			Discord: d,
			TTS:     tts,
			Log:     &provider.MockLogProvider{},
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	node := &CompiledFlowNode{
		ID:   "0",
		Type: FlowNodeTypeEntryEvent,
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{
			{ID: "1", Type: FlowNodeTypeActionTTSSpeak, Data: data},
		}},
	}
	return node.Execute(c)
}

func TestTTSSpeakPlaysInVoiceChannel(t *testing.T) {
	d := newTTSTestDiscord()
	s := &ttsTestSpeaker{}
	err := executeTTSBlock(t, d, s, `{
		"channel_target": "3",
		"audio_leave_after": true,
		"tts_data": {"text": " Hello there ", "voice": "nova", "speed": 1.5}
	}`)
	require.NoError(t, err)

	require.Len(t, s.calls, 1)
	assert.Equal(t, provider.SpeakOpts{Text: "Hello there", Voice: "nova", Speed: 1.5}, s.calls[0])

	require.NotNil(t, d.played)
	assert.Equal(t, discord.GuildID(5), d.played.GuildID)
	assert.Equal(t, discord.ChannelID(3), d.played.ChannelID)
	assert.Equal(t, "sound of Hello there", string(d.played.Sound))
	assert.Empty(t, d.played.AssetID)
	assert.True(t, d.played.LeaveAfter)
}

func TestTTSSpeakDefaults(t *testing.T) {
	d := newTTSTestDiscord()
	s := &ttsTestSpeaker{}
	err := executeTTSBlock(t, d, s, `{"channel_target": "3", "tts_data": {"text": "Hi"}}`)
	require.NoError(t, err)

	require.Len(t, s.calls, 1)
	assert.Equal(t, ttsDefaultVoice, s.calls[0].Voice)
	assert.Zero(t, s.calls[0].Speed)
	require.NotNil(t, d.played)
	assert.False(t, d.played.LeaveAfter)
}

func TestTTSSpeakErrors(t *testing.T) {
	long := strings.Repeat("a", ttsMaxLength+1)

	tests := []struct {
		name    string
		data    string
		tts     provider.TTSProvider
		channel *discord.Channel
		want    string
		// spoken is whether the speech is requested before the error.
		spoken bool
	}{
		{name: "empty text", data: `{"channel_target": "3", "tts_data": {"text": "  "}}`, want: "text is empty"},
		{name: "long text", data: `{"channel_target": "3", "tts_data": {"text": "` + long + `"}}`, want: "more than the limit"},
		{name: "unknown voice", data: `{"channel_target": "3", "tts_data": {"text": "Hi", "voice": "robot"}}`, want: "unsupported voice"},
		{name: "speed too high", data: `{"channel_target": "3", "tts_data": {"text": "Hi", "speed": 5}}`, want: "speed must be between"},
		{name: "no channel", data: `{"tts_data": {"text": "Hi"}}`, want: "set the voice channel"},
		{name: "text channel", data: `{"channel_target": "3", "tts_data": {"text": "Hi"}}`, channel: &discord.Channel{ID: 3, GuildID: 5, Type: discord.GuildText}, want: "is not a voice channel"},
		{name: "stage channel", data: `{"channel_target": "3", "tts_data": {"text": "Hi"}}`, channel: &discord.Channel{ID: 3, GuildID: 5, Type: discord.GuildStageVoice}, want: "stage channels"},
		{name: "no provider", data: `{"channel_target": "3", "tts_data": {"text": "Hi"}}`, tts: nil, want: provider.ErrTTSUnavailable.Error()},
		{name: "unavailable", data: `{"channel_target": "3", "tts_data": {"text": "Hi"}}`, tts: provider.NoTTSProvider{}, want: provider.ErrTTSUnavailable.Error()},
		{name: "provider error", data: `{"channel_target": "3", "tts_data": {"text": "Hi"}}`, tts: &ttsTestSpeaker{err: errors.New("service down")}, want: "service down", spoken: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTTSTestDiscord()
			if tt.channel != nil {
				d.channel = tt.channel
			}

			var speaker *ttsTestSpeaker
			tts := tt.tts
			if tts == nil && tt.name != "no provider" {
				speaker = &ttsTestSpeaker{}
				tts = speaker
			}
			if s, ok := tts.(*ttsTestSpeaker); ok {
				speaker = s
			}

			err := executeTTSBlock(t, d, tts, tt.data)
			assert.ErrorContains(t, err, tt.want)
			assert.Nil(t, d.played, "nothing is played when the speech fails")
			if speaker != nil {
				// Everything that can be checked is checked before the
				// speech is requested, which costs money.
				assert.Equal(t, tt.spoken, len(speaker.calls) > 0)
			}
		})
	}
}

func TestTTSSpeakPlayError(t *testing.T) {
	d := newTTSTestDiscord()
	d.playErr = errors.New("playing audio requires premium")

	err := executeTTSBlock(t, d, &ttsTestSpeaker{}, `{"channel_target": "3", "tts_data": {"text": "Hi"}}`)
	assert.ErrorContains(t, err, "requires premium")
}

func TestTTSDataValidate(t *testing.T) {
	assert.NoError(t, TTSData{Text: "Hi"}.Validate())
	assert.NoError(t, TTSData{Text: "Hi", Voice: "onyx", Speed: ttsMaxSpeed}.Validate())
	assert.NoError(t, TTSData{Text: "Hi", Speed: ttsMinSpeed}.Validate())

	assert.Error(t, TTSData{}.Validate())
	assert.Error(t, TTSData{Text: strings.Repeat("a", ttsMaxLength+1)}.Validate())
	assert.Error(t, TTSData{Text: "Hi", Voice: "robot"}.Validate())
	assert.Error(t, TTSData{Text: "Hi", Speed: 0.1}.Validate())
	assert.Error(t, TTSData{Text: "Hi", Speed: 4.5}.Validate())
}
