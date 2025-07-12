package counting

import (
	"context"

	"github.com/kitecloud/kite/kite-service/pkg/module"
)

const channelsConfigKey = "channels"

type CountingModule struct {
}

func NewCountingModule() *CountingModule {
	return &CountingModule{}
}

func (p *CountingModule) Instance(ctx context.Context, appID string, config module.ConfigValues) (module.ModuleInstance, error) {
	return &CountingModuleInstance{
		appID:  appID,
		config: config,
	}, nil
}

func (p *CountingModule) ID() string {
	return "counting"
}

func (p *CountingModule) IsDefault() bool {
	return false
}

func (p *CountingModule) Metadata() module.Metadata {
	return module.Metadata{
		Name:        "Counting",
		Description: "Create counting channels where users can try to count up.",
		Icon:        "calculator",
		Author:      "Merlin",
	}
}

func (p *CountingModule) Config() module.Config {
	return module.Config{
		Sections: []module.ConfigSection{
			{
				Name:        "Counting",
				Description: "Counting configuration",
				Fields: []module.ConfigField{
					{
						Key:         channelsConfigKey,
						Name:        "Channels",
						Description: "The channels to count messages in",
						Type:        module.ConfigFieldTypeArray,
						ItemType:    module.ConfigFieldTypeString,
					},
				},
			},
		},
	}
}
