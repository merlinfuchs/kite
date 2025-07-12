package counting

import (
	"context"

	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

const channelsConfigKey = "channels"

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
		Icon:        "calculator",
		Author:      "Merlin",
	}
}

func (p *CountingPlugin) Config() plugin.Config {
	return plugin.Config{
		Sections: []plugin.ConfigSection{
			{
				Name:        "Counting",
				Description: "Counting configuration",
				Fields: []plugin.ConfigField{
					{
						Key:         channelsConfigKey,
						Name:        "Channels",
						Description: "The channels to count messages in",
						Type:        plugin.ConfigFieldTypeArray,
						ItemType:    plugin.ConfigFieldTypeString,
					},
				},
			},
		},
	}
}
