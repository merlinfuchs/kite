package engine

import (
	"context"
	"log/slog"
	"slices"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type pluginInstance struct {
	model    *model.PluginInstance
	plugin   plugin.Plugin
	instance plugin.PluginInstance
	env      Env

	eventTypes   map[ws.EventType]bool
	commandNames map[string]bool
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

		eventTypes:   computeEventTypes(plugin, model.EnabledResourceIDs),
		commandNames: computeCommandNames(plugin, model.EnabledResourceIDs),
	}
}

func (p *pluginInstance) Update(ctx context.Context, model *model.PluginInstance) error {
	p.model = model
	p.eventTypes = computeEventTypes(p.plugin, model.EnabledResourceIDs)
	p.commandNames = computeCommandNames(p.plugin, model.EnabledResourceIDs)
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

// WantsEvent reports whether HandleEvent would do any work for this event.
// Component and modal interactions are always dispatched; they are routed by
// their own data type rather than by the plugin's subscribed event types.
//
// A command interaction is only dispatched to the instance whose enabled
// commands own the name. Interactions otherwise reach every plugin instance on
// the app and each plugin matches on the name itself -- which meant an app
// could reach a plugin command it had left disabled just by defining its own
// command of that name.
func (p *pluginInstance) WantsEvent(event gateway.Event) bool {
	if e, ok := event.(*gateway.InteractionCreateEvent); ok {
		if d, ok := e.Data.(*discord.CommandInteraction); ok {
			return p.commandNames[d.Name]
		}

		return true
	}

	return p.eventTypes[event.EventType()]
}

func (p *pluginInstance) HandleEvent(ctx context.Context, session *state.State, event gateway.Event) {
	var err error

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		switch e.Data.(type) {
		case *discord.CommandInteraction:
			err = p.instance.HandleCommand(p.pluginContext(ctx, session), e)
		case discord.ComponentInteraction:
			err = p.instance.HandleComponent(p.pluginContext(ctx, session), e)
		case *discord.ModalInteraction:
			err = p.instance.HandleModal(p.pluginContext(ctx, session), e)
		}
	default:
		err = p.instance.HandleEvent(p.pluginContext(ctx, session), event)
	}

	if err != nil {
		// TODO: Log to the plugin instance's log store
		slog.Error(
			"Failed to handle event from plugin",
			slog.String("plugin_id", p.model.PluginID),
			slog.String("event_type", string(event.EventType())),
			slog.String("error", err.Error()),
		)
	}
}

func (p *pluginInstance) pluginContext(ctx context.Context, session *state.State) *pluginContext {
	return &pluginContext{
		Context:       ctx,
		ValueProvider: NewValueProvider(p.model.ID, p.env.PluginValueStore),
		appID:         p.model.AppID,
		discord:       NewDiscordProvider(p.model.AppID, p.env.AppStore, session),
	}
}

func computeEventTypes(pl plugin.Plugin, enabledResourceIDs []string) map[ws.EventType]bool {
	types := make(map[ws.EventType]bool)
	for _, event := range pl.Events() {
		types[event.Type.DiscordEventType()] = slices.Contains(enabledResourceIDs, event.ID)
	}
	return types
}

// computeCommandNames indexes the Discord command names owned by the enabled
// commands. Computed alongside eventTypes so that dispatch is a map lookup
// rather than a rebuild of the plugin's whole command tree per interaction.
func computeCommandNames(pl plugin.Plugin, enabledResourceIDs []string) map[string]bool {
	names := make(map[string]bool)
	for _, command := range pl.Commands() {
		if slices.Contains(enabledResourceIDs, command.ID) {
			names[command.Data.Name] = true
		}
	}
	return names
}

func (a *App) dispatchEventToPlugins(session *state.State, event gateway.Event) {
	a.RLock()
	defer a.RUnlock()

	for _, plugin := range a.pluginInstances {
		// Check before spawning. This filter used to run inside the goroutine,
		// so every event spawned one goroutine per plugin instance only for
		// most of them to return immediately.
		if !plugin.WantsEvent(event) {
			continue
		}

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
