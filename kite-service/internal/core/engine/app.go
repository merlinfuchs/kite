package engine

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

type App struct {
	sync.RWMutex

	id string

	env Env

	pluginInstances map[string]*pluginInstance
	commands        map[string]*Command
	listeners       map[string]*EventListener
	// TODO?: Cache messages (LRUCache<*MessageInstance>)
}

func NewApp(
	id string,
	stores Env,
) *App {
	return &App{
		id:              id,
		env:             stores,
		commands:        make(map[string]*Command),
		listeners:       make(map[string]*EventListener),
		pluginInstances: make(map[string]*pluginInstance),
	}
}

func (a *App) AddPluginInstance(pluginInstance *model.PluginInstance) {
	plugin := a.env.PluginRegistry.Plugin(pluginInstance.PluginID)
	if plugin == nil {
		slog.Warn(
			"Unknown plugin",
			slog.String("plugin_id", pluginInstance.PluginID),
		)
		return
	}

	a.Lock()
	existing := a.pluginInstances[pluginInstance.ID]
	a.Unlock()

	if existing != nil {
		err := existing.Update(context.TODO(), pluginInstance)
		if err != nil {
			slog.With("error", err).Error("failed to update plugin instance")
			return
		}
	} else {
		instance, err := plugin.Instance(context.TODO(), a.id, pluginInstance.Config)
		if err != nil {
			slog.With("error", err).Error("failed to create module instance")
			return
		}

		a.Lock()
		a.pluginInstances[pluginInstance.ID] = newPluginInstance(
			pluginInstance,
			plugin,
			instance,
			a.env,
		)
		a.Unlock()
	}

	a.Lock()
	defer a.Unlock()
}

func (a *App) RemoveDanglingPluginInstances(pluginInstanceIDs []string) {
	pluginInstanceIDMap := make(map[string]struct{}, len(pluginInstanceIDs))
	for _, pluginInstanceID := range pluginInstanceIDs {
		pluginInstanceIDMap[pluginInstanceID] = struct{}{}
	}

	a.Lock()
	defer a.Unlock()

	for pluginInstanceID, pluginInstance := range a.pluginInstances {
		if _, ok := pluginInstanceIDMap[pluginInstanceID]; !ok {
			err := pluginInstance.Close()
			if err != nil {
				slog.With("error", err).Error("failed to close plugin instance")
			}

			delete(a.pluginInstances, pluginInstanceID)
		}
	}
}

func (a *App) AddCommand(cmd *model.Command) {
	command, err := NewCommand(
		cmd,
		a.env,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create command")
		return
	}

	a.Lock()
	defer a.Unlock()

	a.commands[cmd.ID] = command
}

func (a *App) RemoveDanglingCommands(commandIDs []string) {
	commandIDMap := make(map[string]struct{}, len(commandIDs))
	for _, commandID := range commandIDs {
		commandIDMap[commandID] = struct{}{}
	}

	a.Lock()
	defer a.Unlock()

	for cmdID := range a.commands {
		if _, ok := commandIDMap[cmdID]; !ok {
			delete(a.commands, cmdID)
		}
	}
}

func (a *App) AddEventListener(listener *model.EventListener) {
	eventListener, err := NewEventListener(
		listener,
		a.env,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create event listener")
		return
	}

	a.Lock()
	defer a.Unlock()

	a.listeners[listener.ID] = eventListener
}

func (a *App) RemoveDanglingEventListeners(listenerIDs []string) {
	listenerIDMap := make(map[string]struct{}, len(listenerIDs))
	for _, listenerID := range listenerIDs {
		listenerIDMap[listenerID] = struct{}{}
	}

	a.Lock()
	defer a.Unlock()

	for listenerID := range a.listeners {
		if _, ok := listenerIDMap[listenerID]; !ok {
			delete(a.listeners, listenerID)
		}
	}
}

func (a *App) HandleEvent(appID string, session *state.State, event gateway.Event) {
	a.dispatchEventToPlugins(session, event)

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		timeDiff := time.Since(e.ID.Time())
		if timeDiff > 500*time.Millisecond {
			slog.Warn(
				"Received interaction event late",
				slog.String("app_id", appID),
				slog.String("interaction_id", e.ID.String()),
				slog.String("time_diff", timeDiff.String()),
			)
		}

		switch d := e.Data.(type) {
		case *discord.CommandInteraction:
			fullName := getFullCommandName(d)

			lockStart := time.Now()
			a.RLock()
			defer a.RUnlock()
			lockDiff := time.Since(lockStart)
			if lockDiff > 100*time.Millisecond {
				slog.Warn(
					"Locking app took too long",
					slog.String("app_id", appID),
					slog.String("lock_duration", lockDiff.String()),
				)
			}

			for _, command := range a.commands {
				if command.cmd.Name == fullName {
					go command.HandleEvent(appID, session, event)
					break
				}
			}
		case *discord.ButtonInteraction:
			customID := string(d.CustomID)
			resumePointID, _, isResume := message.DecodeCustomIDMessageComponentResumePoint(customID)
			if isResume {
				resumePoint, err := a.env.ResumePointStore.ResumePoint(context.TODO(), resumePointID)
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
					a.RLock()
					defer a.RUnlock()

					command, ok := a.commands[resumePoint.CommandID.String]
					if !ok {
						return
					}

					node := command.flow.FindChildWithID(resumePoint.FlowNodeID, true)
					if node == nil {
						slog.Error(
							"Failed to find node in flow",
							slog.String("resume_point_id", resumePointID),
							slog.String("command_id", resumePoint.CommandID.String),
						)
						return
					}

					go a.env.executeFlowEvent(
						context.Background(),
						a.id,
						node,
						session,
						event,
						entityLinks{
							CommandID: null.NewString(command.cmd.ID, true),
						},
						&resumePoint.FlowState,
					)
				}
				return
			}

			messageID := e.Message.ID.String()
			messageInstnace, err := a.env.MessageInstanceStore.MessageInstanceByDiscordMessageID(context.TODO(), messageID)
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
				a.env,
			)
			if err != nil {
				slog.With("error", err).Error("failed to create message instance")
				return
			}

			go instance.HandleEvent(appID, session, event)
		case *discord.ModalInteraction:
			customID := string(d.CustomID)
			resumePointID, ok := message.DecodeCustomIDModalResumePoint(customID)
			if !ok {
				return
			}

			resumePoint, err := a.env.ResumePointStore.ResumePoint(context.TODO(), resumePointID)
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
				a.RLock()
				defer a.RUnlock()

				command, ok := a.commands[resumePoint.CommandID.String]
				if !ok {
					return
				}

				node := command.flow.FindChildWithID(resumePoint.FlowNodeID, true)
				if node == nil {
					slog.Error(
						"Failed to find node in flow",
						slog.String("resume_point_id", resumePointID),
						slog.String("command_id", resumePoint.CommandID.String),
					)
					return
				}

				go a.env.executeFlowEvent(
					context.Background(),
					a.id,
					node,
					session,
					event,
					entityLinks{
						CommandID: null.NewString(command.cmd.ID, true),
					},
					&resumePoint.FlowState,
				)
			}

			if resumePoint.MessageInstanceID.Valid {
				messageInstance, err := a.env.MessageInstanceStore.MessageInstance(
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
					a.env,
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

				node := targetFlow.FindChildWithID(resumePoint.FlowNodeID, true)
				if node == nil {
					slog.Error(
						"Failed to find node in flow",
						slog.String("resume_point_id", resumePointID),
						slog.String("message_id", resumePoint.MessageID.String),
						slog.Int64("message_instance_id", resumePoint.MessageInstanceID.Int64),
						slog.String("flow_source_id", resumePoint.FlowSourceID.String),
					)
					return
				}

				go a.env.executeFlowEvent(
					context.Background(),
					a.id,
					node,
					session,
					event,
					entityLinks{},
					&resumePoint.FlowState,
				)
			}
		}
	default:
		eventType := model.EventTypeFromDiscordEventType(e.EventType())

		a.RLock()
		defer a.RUnlock()

		for _, listener := range a.listeners {
			if listener.listener.Source != model.EventSourceDiscord {
				continue
			}

			if listener.listener.Type != eventType {
				continue
			}

			go listener.HandleEvent(appID, session, event)
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
