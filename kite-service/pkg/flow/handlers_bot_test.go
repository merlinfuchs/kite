package flow

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type audioTestDiscordProvider struct {
	provider.MockDiscordProvider

	channel discord.Channel
	played  *provider.PlayAudioOpts
	stopped discord.GuildID
}

func (p *audioTestDiscordProvider) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	return &p.channel, nil
}

func (p *audioTestDiscordProvider) PlayAudio(ctx context.Context, opts provider.PlayAudioOpts) error {
	p.played = &opts
	return nil
}

func (p *audioTestDiscordProvider) StopAudio(ctx context.Context, guildID discord.GuildID) error {
	p.stopped = guildID
	return nil
}

type audioTestHTTPProvider struct {
	status        int
	body          []byte
	contentLength int64
	url           string
}

func (p *audioTestHTTPProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	p.url = req.URL.String()
	return &http.Response{
		StatusCode:    p.status,
		Body:          io.NopCloser(bytes.NewReader(p.body)),
		ContentLength: p.contentLength,
	}, nil
}

func (p *audioTestHTTPProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	return p.HTTPRequest(ctx, req)
}

func executeAudioBlock(t *testing.T, d *audioTestDiscordProvider, h *audioTestHTTPProvider, nodeType FlowNodeType, dataJSON string) error {
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
			HTTP:    h,
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
			{ID: "1", Type: nodeType, Data: data},
		}},
	}
	return node.Execute(c)
}

func newAudioTestProviders() (*audioTestDiscordProvider, *audioTestHTTPProvider) {
	d := &audioTestDiscordProvider{channel: discord.Channel{ID: 3, GuildID: 7, Type: discord.GuildVoice}}
	h := &audioTestHTTPProvider{status: http.StatusOK, body: []byte("audio"), contentLength: -1}
	return d, h
}

func TestVoiceAudioPlay(t *testing.T) {
	d, h := newAudioTestProviders()
	err := executeAudioBlock(t, d, h, FlowNodeTypeActionVoiceAudioPlay, `{
		"channel_target": "3",
		"audio_url": "https://cdn.example.com/sound.mp3",
		"audio_volume": "50",
		"audio_leave_after": true
	}`)
	require.NoError(t, err)

	assert.Equal(t, "https://cdn.example.com/sound.mp3", h.url)
	require.NotNil(t, d.played)
	assert.Equal(t, provider.PlayAudioOpts{
		GuildID:    7,
		ChannelID:  3,
		Audio:      []byte("audio"),
		Volume:     0.5,
		LeaveAfter: true,
	}, *d.played)
}

func TestVoiceAudioPlayDefaultVolume(t *testing.T) {
	d, h := newAudioTestProviders()
	err := executeAudioBlock(t, d, h, FlowNodeTypeActionVoiceAudioPlay, `{
		"channel_target": "3",
		"audio_url": "https://cdn.example.com/sound.mp3"
	}`)
	require.NoError(t, err)
	require.NotNil(t, d.played)
	assert.Equal(t, 1.0, d.played.Volume)
	assert.False(t, d.played.LeaveAfter)
}

func TestVoiceAudioPlayErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		prepare func(d *audioTestDiscordProvider, h *audioTestHTTPProvider)
		err     string
	}{
		{
			name: "not a voice channel",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3"}`,
			prepare: func(d *audioTestDiscordProvider, h *audioTestHTTPProvider) {
				d.channel.Type = discord.GuildText
			},
			err: "not a voice channel",
		},
		{
			name: "stage channel",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3"}`,
			prepare: func(d *audioTestDiscordProvider, h *audioTestHTTPProvider) {
				d.channel.Type = discord.GuildStageVoice
			},
			err: "stage channels",
		},
		{
			name: "no URL",
			data: `{"channel_target": "3"}`,
			err:  "must be a link",
		},
		{
			name: "not http",
			data: `{"channel_target": "3", "audio_url": "file:///etc/passwd"}`,
			err:  "must be a link",
		},
		{
			name: "invalid volume",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3", "audio_volume": "loud"}`,
			err:  "volume must be a number",
		},
		{
			name: "volume too high",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3", "audio_volume": "500"}`,
			err:  "volume must be a number",
		},
		{
			name: "download failed",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3"}`,
			prepare: func(d *audioTestDiscordProvider, h *audioTestHTTPProvider) {
				h.status = http.StatusNotFound
			},
			err: "status 404",
		},
		{
			name: "announced too large",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3"}`,
			prepare: func(d *audioTestDiscordProvider, h *audioTestHTTPProvider) {
				h.contentLength = maxAudioSize + 1
			},
			err: "can't be larger",
		},
		{
			name: "too large",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3"}`,
			prepare: func(d *audioTestDiscordProvider, h *audioTestHTTPProvider) {
				h.body = make([]byte, maxAudioSize+1)
			},
			err: "can't be larger",
		},
		{
			name: "empty file",
			data: `{"channel_target": "3", "audio_url": "https://example.com/a.mp3"}`,
			prepare: func(d *audioTestDiscordProvider, h *audioTestHTTPProvider) {
				h.body = nil
			},
			err: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, h := newAudioTestProviders()
			if tt.prepare != nil {
				tt.prepare(d, h)
			}

			err := executeAudioBlock(t, d, h, FlowNodeTypeActionVoiceAudioPlay, tt.data)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
			assert.Nil(t, d.played, "nothing is played")
		})
	}
}

func TestVoiceAudioStop(t *testing.T) {
	d, h := newAudioTestProviders()
	err := executeAudioBlock(t, d, h, FlowNodeTypeActionVoiceAudioStop, `{}`)
	require.NoError(t, err)
	assert.Equal(t, discord.GuildID(5), d.stopped, "the server the flow runs in")
}
