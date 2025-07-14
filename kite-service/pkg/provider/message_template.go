package provider

import (
	"context"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

type MessageTemplateInstance struct {
	MessageTemplateID string
	MessageID         discord.MessageID
	ChannelID         discord.ChannelID
	GuildID           discord.GuildID
	Ephemeral         bool
}

// MessageTemplateProvider provides access to retrieving message templates and linking them to message instances.
type MessageTemplateProvider interface {
	MessageTemplate(ctx context.Context, id string) (*message.MessageData, error)
	LinkMessageTemplateInstance(ctx context.Context, instance MessageTemplateInstance) error
}

type MockMessageTemplateProvider struct{}

func (p *MockMessageTemplateProvider) MessageTemplate(ctx context.Context, id string) (*message.MessageData, error) {
	return nil, nil
}

func (p *MockMessageTemplateProvider) LinkMessageTemplateInstance(ctx context.Context, instance MessageTemplateInstance) error {
	return nil
}
