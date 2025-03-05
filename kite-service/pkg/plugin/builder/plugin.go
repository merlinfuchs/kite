package builder

import (
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

const PluginID = "builder"

// BuilderPlugin encapsulates the logic for custom commands, event listeners, and message templates.
// It forwards events to the correctl flow and deploys command changes to Discord.
type BuilderPlugin struct {
	env Env
}

func NewBuilderPlugin(
	env Env,
) *BuilderPlugin {
	return &BuilderPlugin{
		env: env,
	}
}

func (p *BuilderPlugin) Instance(appID string, config plugin.ConfigValues) (plugin.PluginInstance, error) {
	return newBuilderPluginInstance(appID, config, p.env), nil
}

func (p *BuilderPlugin) ID() string {
	return PluginID
}

func (p *BuilderPlugin) IsDefault() bool {
	return true
}

func (p *BuilderPlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "Builder",
		Description: "Allows you to make your own commands, event listeners, and interactive messages.",
		Icon:        "workflow",
		Author:      "Merlin",
	}
}

func (p *BuilderPlugin) Config() plugin.Config {
	return plugin.Config{
		Sections: []plugin.ConfigSection{},
	}
}
