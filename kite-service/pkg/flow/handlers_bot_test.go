package flow

import (
	"context"
	"encoding/json"
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

func executeAudioBlock(t *testing.T, d *audioTestDiscordProvider, nodeType FlowNodeType, dataJSON string) error {
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

func newAudioTestProvider() *audioTestDiscordProvider {
	return &audioTestDiscordProvider{channel: discord.Channel{ID: 3, GuildID: 7, Type: discord.GuildVoice}}
}

func TestVoiceAudioPlay(t *testing.T) {
	d := newAudioTestProvider()
	err := executeAudioBlock(t, d, FlowNodeTypeActionVoiceAudioPlay, `{
		"channel_target": "3",
		"audio_asset_id": "sound",
		"audio_leave_after": true
	}`)
	require.NoError(t, err)

	require.NotNil(t, d.played)
	assert.Equal(t, provider.PlayAudioOpts{
		GuildID:    7,
		ChannelID:  3,
		AssetID:    "sound",
		LeaveAfter: true,
	}, *d.played)
}

func TestVoiceAudioPlayErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		prepare func(d *audioTestDiscordProvider)
		err     string
	}{
		{
			name: "not a voice channel",
			data: `{"channel_target": "3", "audio_asset_id": "sound"}`,
			prepare: func(d *audioTestDiscordProvider) {
				d.channel.Type = discord.GuildText
			},
			err: "not a voice channel",
		},
		{
			name: "stage channel",
			data: `{"channel_target": "3", "audio_asset_id": "sound"}`,
			prepare: func(d *audioTestDiscordProvider) {
				d.channel.Type = discord.GuildStageVoice
			},
			err: "stage channels",
		},
		{
			name: "no sound",
			data: `{"channel_target": "3"}`,
			err:  "upload a sound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newAudioTestProvider()
			if tt.prepare != nil {
				tt.prepare(d)
			}

			err := executeAudioBlock(t, d, FlowNodeTypeActionVoiceAudioPlay, tt.data)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
			assert.Nil(t, d.played, "nothing is played")
		})
	}
}

func TestVoiceAudioStop(t *testing.T) {
	d := newAudioTestProvider()
	err := executeAudioBlock(t, d, FlowNodeTypeActionVoiceAudioStop, `{}`)
	require.NoError(t, err)
	assert.Equal(t, discord.GuildID(5), d.stopped, "the server the flow runs in")
}
