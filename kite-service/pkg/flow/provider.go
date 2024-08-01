package flow

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

type FlowProviders struct {
	Discord FlowDiscordProvider
	KV      FlowKVProvider
	HTTP    FlowHTTPProvider
	Log     FlowLogProvider
}

type FlowDiscordProvider interface {
	CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) error
	CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error)
}

type FlowKVProvider interface{}

type FlowHTTPProvider interface{}

type FlowLogProvider interface {
	CreateLogEntry(ctx context.Context, level LogLevel, message string)
}
