package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memberEditDiscordProvider struct {
	provider.MockDiscordProvider

	calls []api.ModifyMemberData
}

func (p *memberEditDiscordProvider) EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error {
	p.calls = append(p.calls, data)
	return nil
}

func executeMemberVoiceEdit(t *testing.T, data FlowNodeData) (*memberEditDiscordProvider, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &memberEditDiscordProvider{}

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{
			Discord: discordProvider,
			Log:     &provider.MockLogProvider{},
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	defer c.Cancel()

	data.UserTarget = "123"
	node := &CompiledFlowNode{
		ID:   "1",
		Type: FlowNodeTypeActionMemberVoiceEdit,
		Data: data,
	}

	return discordProvider, node.Execute(c)
}

func TestMemberVoiceEditOnlySendsChangedFields(t *testing.T) {
	p, err := executeMemberVoiceEdit(t, FlowNodeData{
		MemberVoiceMute: VoiceStateChangeOn,
		MemberVoiceDeaf: VoiceStateChangeUnchanged,
	})
	require.NoError(t, err)
	require.Len(t, p.calls, 1)

	call := p.calls[0]
	require.NotNil(t, call.Mute)
	assert.True(t, *call.Mute)
	assert.Nil(t, call.Deaf)
	assert.Equal(t, discord.ChannelID(0), call.VoiceChannel)
}

func TestMemberVoiceEditMovesAndUndeafens(t *testing.T) {
	p, err := executeMemberVoiceEdit(t, FlowNodeData{
		MemberVoiceDeaf: VoiceStateChangeOff,
		ChannelTarget:   "456",
	})
	require.NoError(t, err)
	require.Len(t, p.calls, 1)

	call := p.calls[0]
	assert.Nil(t, call.Mute)
	require.NotNil(t, call.Deaf)
	assert.False(t, *call.Deaf)
	assert.Equal(t, discord.ChannelID(456), call.VoiceChannel)
}

func TestMemberVoiceEditRejectsInvalidChannel(t *testing.T) {
	for _, target := range []string{"not-a-channel", "0", "-1", "{{ '' }}"} {
		t.Run(target, func(t *testing.T) {
			p, err := executeMemberVoiceEdit(t, FlowNodeData{
				MemberVoiceMute: VoiceStateChangeOn,
				ChannelTarget:   target,
			})
			require.ErrorContains(t, err, "not a valid channel ID")
			assert.Empty(t, p.calls)
		})
	}
}

func TestMemberVoiceEditRejectsNoChanges(t *testing.T) {
	p, err := executeMemberVoiceEdit(t, FlowNodeData{
		MemberVoiceMute: VoiceStateChangeUnchanged,
	})
	require.ErrorContains(t, err, "nothing to change")
	assert.Empty(t, p.calls)
}

func TestMemberVoiceEditRejectsUnknownChange(t *testing.T) {
	p, err := executeMemberVoiceEdit(t, FlowNodeData{
		MemberVoiceMute: "maybe",
	})
	require.Error(t, err)
	assert.Empty(t, p.calls)

	err = FlowNodeData{MemberVoiceDeaf: "maybe"}.Validate(FlowNodeTypeActionMemberVoiceEdit)
	assert.Error(t, err)
}
