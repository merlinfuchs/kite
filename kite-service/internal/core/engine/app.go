package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

type App struct {
	sync.RWMutex

	id string

	stores               Env
	hasUndeployedChanges bool

	commands map[string]*Command
	// TODO?: Cache messages (LRUCache<*MessageInstance>)
	listeners map[string]*EventListener
	plugins   map[string]*PluginInstance
}

func NewApp(
	id string,
	stores Env,
) *App {
	return &App{
		id:        id,
		stores:    stores,
		commands:  make(map[string]*Command),
		listeners: make(map[string]*EventListener),
		plugins:   make(map[string]*PluginInstance),
	}
}

func (a *App) AddCommand(cmd *model.Command) {
	a.Lock()
	defer a.Unlock()

	command, err := NewCommand(
		cmd,
		a.stores,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create command")
		return
	}

	a.commands[cmd.ID] = command

	if !cmd.LastDeployedAt.Valid || cmd.LastDeployedAt.Time.Before(cmd.UpdatedAt) {
		a.hasUndeployedChanges = true
	}
}

func (a *App) RemoveDanglingCommands(commandIDs []string) {
	a.Lock()
	defer a.Unlock()

	commandIDMap := make(map[string]struct{}, len(commandIDs))
	for _, commandID := range commandIDs {
		commandIDMap[commandID] = struct{}{}
	}

	for cmdID := range a.commands {
		if _, ok := commandIDMap[cmdID]; !ok {
			delete(a.commands, cmdID)
			a.hasUndeployedChanges = true
		}
	}
}

func (a *App) AddEventListener(listener *model.EventListener) {
	a.Lock()
	defer a.Unlock()

	eventListener, err := NewEventListener(
		listener,
		a.stores,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create event listener")
		return
	}

	a.listeners[listener.ID] = eventListener
}

func (a *App) RemoveDanglingEventListeners(listenerIDs []string) {
	a.Lock()
	defer a.Unlock()

	listenerIDMap := make(map[string]struct{}, len(listenerIDs))
	for _, listenerID := range listenerIDs {
		listenerIDMap[listenerID] = struct{}{}
	}

	for listenerID := range a.listeners {
		if _, ok := listenerIDMap[listenerID]; !ok {
			delete(a.listeners, listenerID)
		}
	}
}

func (a *App) AddPlugin(plugin *model.PluginInstance) {
	a.Lock()
	defer a.Unlock()

	pluginInstance, err := NewPluginInstance(a.id, plugin, a.stores)
	if err != nil {
		a.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to create plugin instance: %s", err))
		return
	}

	a.plugins[plugin.PluginID] = pluginInstance
}

func (a *App) RemoveDanglingPlugins(pluginIDs []string) {
	a.Lock()
	defer a.Unlock()

	pluginIDMap := make(map[string]struct{}, len(pluginIDs))
	for _, pluginID := range pluginIDs {
		pluginIDMap[pluginID] = struct{}{}
	}

	for pluginID := range a.plugins {
		if _, ok := pluginIDMap[pluginID]; !ok {
			delete(a.plugins, pluginID)
		}
	}
}

func (a *App) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := a.stores.LogStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     a.id,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", a.id).Error("Failed to create log entry from engine app")
	}
}

func (a *App) HandleEvent(appID string, session *state.State, event gateway.Event) {
	a.RLock()
	defer a.RUnlock()

	for _, plugin := range a.plugins {
		// TODO: check if plugin is interested in this event
		plugin.HandleEvent(appID, session, event)
	}

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		switch d := e.Data.(type) {
		case *discord.CommandInteraction:
			fullName := getFullCommandName(d)
			for _, command := range a.commands {
				if command.cmd.Name == fullName {
					go command.HandleEvent(appID, session, event)
				}
			}
		case *discord.ButtonInteraction:
			messageID := e.Message.ID.String()
			messageInstnace, err := a.stores.MessageInstanceStore.MessageInstanceByDiscordMessageID(context.TODO(), messageID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return
				}

				slog.With("error", err).Error("failed to get message instance by discord message ID")
				return
			}

			instance, err := NewMessageInstance(
				a.id,
				messageInstnace,
				a.stores,
			)
			if err != nil {
				slog.With("error", err).Error("failed to create message instance")
				return
			}

			go instance.HandleEvent(appID, session, event)
		case *discord.ModalInteraction:
			customID := string(d.CustomID)
			if !strings.HasPrefix(customID, "resume:") {
				return
			}

			resumePointID := customID[len("resume:"):]
			resumePoint, err := a.stores.ResumePointStore.ResumePoint(context.TODO(), resumePointID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return
				}

				slog.Error(
					"Failed to get resume point",
					slog.String("resume_point_id", resumePointID),
					slog.String("error", err.Error()),
				)
				return
			}

			if resumePoint.CommandID.Valid {
				command, ok := a.commands[resumePoint.CommandID.String]
				if !ok {
					return
				}

				node := command.flow.FindChildWithID(resumePoint.FlowNodeID)

				a.stores.executeFlowEvent(
					a.id,
					node,
					session,
					event,
					entityLinks{
						CommandID: null.NewString(command.cmd.ID, true),
					},
					&resumePoint.FlowState,
					true,
				)
			}

			if resumePoint.MessageInstanceID.Valid {
				messageInstance, err := a.stores.MessageInstanceStore.MessageInstance(
					context.TODO(),
					resumePoint.MessageID.String,
					uint64(resumePoint.MessageInstanceID.Int64),
				)
				if err != nil {
					if !errors.Is(err, store.ErrNotFound) {
						slog.Error(
							"Failed to get message instance from resume point",
							slog.String("resume_point_id", resumePointID),
							slog.String("message_id", resumePoint.MessageID.String),
							slog.Int64("message_instance_id", resumePoint.MessageInstanceID.Int64),
							slog.String("error", err.Error()),
						)
					}
					return
				}

				instance, err := NewMessageInstance(
					a.id,
					messageInstance,
					a.stores,
				)
				if err != nil {
					slog.Error(
						"Failed to create message instance",
						slog.String("resume_point_id", resumePointID),
						slog.String("message_id", resumePoint.MessageID.String),
						slog.Int64("message_instance_id", resumePoint.MessageInstanceID.Int64),
						slog.String("error", err.Error()),
					)
					return
				}

				targetFlow, ok := instance.flows[resumePoint.FlowSourceID.String]
				if !ok {
					slog.Error(
						"Failed to get target flow from resume point",
						slog.String("resume_point_id", resumePointID),
						slog.String("message_id", resumePoint.MessageID.String),
						slog.Int64("message_instance_id", resumePoint.MessageInstanceID.Int64),
						slog.String("flow_source_id", resumePoint.FlowSourceID.String),
					)
					return
				}

				node := targetFlow.FindChildWithID(resumePoint.FlowNodeID)

				a.stores.executeFlowEvent(
					a.id,
					node,
					session,
					event,
					entityLinks{},
					&resumePoint.FlowState,
					true,
				)
			}
		}
	default:
		eventType := model.EventTypeFromDiscordEventType(e.EventType())
		for _, listener := range a.listeners {
			if listener.listener.Source != model.EventSourceDiscord {
				continue
			}

			if listener.listener.Type != eventType {
				continue
			}

			listener.HandleEvent(appID, session, event)
		}
	}
}

func getFullCommandName(d *discord.CommandInteraction) string {
	fullName := d.Name
	for _, option := range d.Options {
		if option.Type == discord.SubcommandOptionType {
			fullName += " " + option.Name
			break
		} else if option.Type == discord.SubcommandGroupOptionType {
			fullName += " " + option.Name
			for _, subOption := range option.Options {
				fullName += " " + subOption.Name
			}
			break
		}
	}

	return fullName
}
