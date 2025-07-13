package engine

import (
	"context"
	"log/slog"
	"slices"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type pluginInstance struct {
	model    *model.PluginInstance
	plugin   plugin.Plugin
	instance plugin.PluginInstance
	env      Env
}

func newPluginInstance(
	model *model.PluginInstance,
	plugin plugin.Plugin,
	instance plugin.PluginInstance,
	env Env,
) *pluginInstance {
	return &pluginInstance{
		model:    model,
		plugin:   plugin,
		instance: instance,
		env:      env,
	}
}

func (p *pluginInstance) Update(ctx context.Context, model *model.PluginInstance) error {
	p.model = model
	return p.instance.Update(ctx, model.Config)
}

func (p *pluginInstance) Close() error {
	return p.instance.Close()
}

func (p *pluginInstance) Commands() []plugin.Command {
	commands := p.plugin.Commands()

	filtered := make([]plugin.Command, 0, len(commands))
	for _, command := range commands {
		if slices.Contains(p.model.EnabledResourceIDs, command.ID) {
			filtered = append(filtered, command)
		}
	}

	return filtered
}

func (p *pluginInstance) Events() []plugin.Event {
	events := p.plugin.Events()

	filtered := make([]plugin.Event, 0, len(events))
	for _, event := range events {
		if slices.Contains(p.model.EnabledResourceIDs, event.ID) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

func (p *pluginInstance) HandleEvent(ctx context.Context, session *state.State, event gateway.Event) {
	pluginCtx := &pluginContext{
		Context:       context.TODO(),
		ValueProvider: NewValueProvider(p.model.ID, p.env.PluginValueStore),
		appID:         p.model.AppID,
		discord:       NewDiscordProvider(p.model.AppID, p.env.AppStore, session),
	}

	err := p.instance.HandleEvent(pluginCtx, event)
	if err != nil {
		slog.Error(
			"Failed to handle event from plugin",
			slog.String("plugin_id", p.model.PluginID),
			slog.String("error", err.Error()),
		)
		return
	}

}

func (a *App) dispatchEventToPlugins(session *state.State, event gateway.Event) {
	a.RLock()
	defer a.RUnlock()

	for _, plugin := range a.pluginInstances {
		go plugin.HandleEvent(context.TODO(), session, event)
	}
}

type pluginContext struct {
	context.Context

	provider.ValueProvider

	appID   string
	discord provider.DiscordProvider
}

func (c *pluginContext) Discord() provider.DiscordProvider {
	return c.discord
}
