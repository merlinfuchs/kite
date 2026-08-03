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

	eventTypes map[ws.EventType]bool
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

		eventTypes: computeEventTypes(plugin, model.EnabledResourceIDs),
	}
}

func (p *pluginInstance) Update(ctx context.Context, model *model.PluginInstance) error {
	p.model = model
	p.eventTypes = computeEventTypes(p.plugin, model.EnabledResourceIDs)
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

// handlesCommand reports whether the given Discord command name belongs to one
// of this instance's enabled commands. Commands() is already filtered by
// EnabledResourceIDs, so a command the app disabled is not matched.
func (p *pluginInstance) handlesCommand(name string) bool {
	for _, command := range p.Commands() {
		if command.Data.Name == name {
			return true
		}
	}

	return false
}

// WantsEvent reports whether HandleEvent would do any work for this event.
// Interactions are always dispatched; they are routed by their own data type
// rather than by the plugin's subscribed event types.
func (p *pluginInstance) WantsEvent(event gateway.Event) bool {
	if _, ok := event.(*gateway.InteractionCreateEvent); ok {
		return true
	}

	return p.eventTypes[event.EventType()]
}

func (p *pluginInstance) HandleEvent(ctx context.Context, session *state.State, event gateway.Event) {
	var err error

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		switch d := e.Data.(type) {
		case *discord.CommandInteraction:
			// Interactions are dispatched to every plugin instance on the app,
			// so without this each plugin sees every command and matches on the
			// name itself -- which meant an app could reach a plugin command it
			// had left disabled just by defining its own command of that name.
			if !p.handlesCommand(d.Name) {
				return
			}

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
