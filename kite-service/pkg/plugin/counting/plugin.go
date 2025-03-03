package counting

import (
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

const channelsConfigKey = "channels"

type CountingPlugin struct{}

func NewCountingPlugin() *CountingPlugin {
	return &CountingPlugin{}
}

func (p *CountingPlugin) Instance(config plugin.ConfigValues) (plugin.PluginInstance, error) {
	return &CountingPluginInstance{
		config: config,
	}, nil
}

func (p *CountingPlugin) ID() string {
	return "counting"
}

func (p *CountingPlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "Counting",
		Description: "A plugin for counting messages in a channel",
		Icon:        "diff",
		Author:      "Merlin",
	}
}

func (p *CountingPlugin) Version() string {
	return "0.0.1"
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
