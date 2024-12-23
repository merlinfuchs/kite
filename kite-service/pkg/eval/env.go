package eval

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"gopkg.in/guregu/null.v4"
)

type MessageTemplateEnv struct {
	VariableEnv

	Interaction *InteractionEnv `expr:"interaction"`
	Event       *EventEnv       `expr:"event"`
}

type FlowEnv struct {
	VariableEnv

	Interaction *InteractionEnv `expr:"interaction"`
	Event       *EventEnv       `expr:"event"`
}

type VariableEnv struct {
	provider VariableProvider

	VariableFunc func(ctx context.Context, id string, scope null.String) (string, error) `expr:"var"`
}

func NewVariableEnv(provider VariableProvider) VariableEnv {
	return VariableEnv{
		provider: provider,
		// VariableFunc: provider.VariableValue,
	}
}

type InteractionEnv struct {
	interaction *discord.InteractionEvent

	ID        string `expr:"id"`
	ChannelID string `expr:"channel_id"`
	GuildID   string `expr:"guild_id"`
}

func NewInteractionEnv(i *discord.InteractionEvent) *InteractionEnv {
	return &InteractionEnv{
		interaction: i,

		ID:        i.ID.String(),
		ChannelID: i.ChannelID.String(),
		GuildID:   i.GuildID.String(),
	}
}

type EventEnv struct {
	event ws.Event

	User    *UserEnv    `expr:"user"`
	Member  *MemberEnv  `expr:"member"`
	Channel *ChannelEnv `expr:"channel"`
	Message *MessageEnv `expr:"message"`
	Guild   *GuildEnv   `expr:"guild"`
}

func NewEventEnv(event ws.Event) *EventEnv {
	return &EventEnv{
		event: event,
	}
}

type UserEnv struct{}

type MemberEnv struct{}

type ChannelEnv struct{}

type MessageEnv struct{}

type GuildEnv struct{}
