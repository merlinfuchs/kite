package counting

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type CountingPlugin struct {
}

func NewCountingPlugin() *CountingPlugin {
	return &CountingPlugin{}
}

func (p *CountingPlugin) Instance(ctx context.Context, appID string, config plugin.ConfigValues) (plugin.PluginInstance, error) {
	return &CountingPluginInstance{
		appID:  appID,
		config: config,
	}, nil
}

func (p *CountingPlugin) ID() string {
	return "counting"
}

func (p *CountingPlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "Counting",
		Description: "Create counting channels where users can try to count up.",
		Icon:        "arrow-up-1-0",
		Author:      "Merlin",
	}
}

func (p *CountingPlugin) Config() plugin.Config {
	return plugin.Config{}
}

func (p *CountingPlugin) Events() []plugin.Event {
	return []plugin.Event{
		{
			ID:          "event_message_create",
			Source:      plugin.EventSourceDiscord,
			Type:        plugin.EventTypeMessageCreate,
			Description: "Check if the message is a counting message and if it is, increment the counter",
		},
	}
}

func (p *CountingPlugin) Commands() []plugin.Command {
	perms := discord.PermissionManageChannels

	return []plugin.Command{
		{
			ID: "cmd_toggle",
			Data: api.CreateCommandData{
				Name:                     "counting-toggle",
				Description:              "Toggle the counting game in the current channel",
				DefaultMemberPermissions: &perms,
			},
		},
	}
}
