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
	"github.com/kitecloud/kite/kite-service/internal/metrics"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

type App struct {
	sync.RWMutex

	id string

	env Env

	pluginInstances map[string]*pluginInstance
	commands        map[string]*Command
	listeners       map[string]*EventListener

	// Lookup indexes derived from commands and listeners. Dispatch used to
	// linear-scan both on every event. Rebuilt wholesale on mutation rather
	// than patched, so renames and type changes can't leave stale entries,
	// and always replaced rather than mutated in place so readers can take a
	// reference and drop the lock before dispatching.
	commandsByName  map[string]*Command
	listenersByType map[model.EventListenerType][]*EventListener
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
		commandsByName:  make(map[string]*Command),
		listenersByType: make(map[model.EventListenerType][]*EventListener),
	}
}

// rebuildCommandIndex regenerates the name lookup from a.commands. Callers
// must hold the write lock.
func (a *App) rebuildCommandIndex() {
	index := make(map[string]*Command, len(a.commands))
	for _, command := range a.commands {
		index[command.cmd.Name] = command
	}
	a.commandsByName = index
}

// rebuildListenerIndex regenerates the event type lookup from a.listeners,
// dropping any listener that isn't sourced from Discord. Callers must hold the
// write lock.
func (a *App) rebuildListenerIndex() {
	index := make(map[model.EventListenerType][]*EventListener, len(a.listeners))
	for _, listener := range a.listeners {
		if listener.listener.Source != model.EventSourceDiscord {
			continue
		}
		index[listener.listener.Type] = append(index[listener.listener.Type], listener)
	}
	a.listenersByType = index
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
}

// RemoveDanglingPluginInstances drops instances absent from enabledIDs, the
// set of plugin instances that still exist and are enabled.
func (a *App) RemoveDanglingPluginInstances(enabledIDs map[string]struct{}) {
	a.Lock()
	defer a.Unlock()

	for pluginInstanceID, pluginInstance := range a.pluginInstances {
		if _, ok := enabledIDs[pluginInstanceID]; !ok {
			err := pluginInstance.Close()
			if err != nil {
				slog.With("error", err).Error("failed to close plugin instance")
			}

			delete(a.pluginInstances, pluginInstanceID)
		}
	}
}

// AddCommand registers an already-compiled command. Compilation happens in the
// caller so it stays off the engine's registry lock.
func (a *App) AddCommand(commandID string, command *Command) {
	lockStart := time.Now()
	a.Lock()
	defer a.Unlock()

	lockDiff := time.Since(lockStart)
	metrics.ObserveLockWait("app_write", lockDiff)

	if lockDiff > 500*time.Millisecond {
		slog.Warn(
			"Locking app for adding command took too long",
			slog.String("app_id", a.id),
			slog.String("lock_duration", lockDiff.String()),
		)
	}

	a.commands[commandID] = command
	a.rebuildCommandIndex()
}

// RemoveDanglingCommands drops commands absent from enabledIDs, the set of
// commands that still exist and are enabled.
func (a *App) RemoveDanglingCommands(enabledIDs map[string]struct{}) {
	a.Lock()
	defer a.Unlock()

	var removed bool
	for cmdID := range a.commands {
		if _, ok := enabledIDs[cmdID]; !ok {
			delete(a.commands, cmdID)
			removed = true
		}
	}

	if removed {
		a.rebuildCommandIndex()
	}
}

// AddEventListener registers an already-compiled event listener. Compilation
// happens in the caller so it stays off the engine's registry lock.
func (a *App) AddEventListener(listenerID string, listener *EventListener) {
	a.Lock()
	defer a.Unlock()

	a.listeners[listenerID] = listener
	a.rebuildListenerIndex()
}

// RemoveDanglingEventListeners drops listeners absent from enabledIDs, the set
// of listeners that still exist and are enabled.
func (a *App) RemoveDanglingEventListeners(enabledIDs map[string]struct{}) {
	a.Lock()
	defer a.Unlock()

	var removed bool
	for listenerID := range a.listeners {
		if _, ok := enabledIDs[listenerID]; !ok {
			delete(a.listeners, listenerID)
			removed = true
		}
	}

	if removed {
		a.rebuildListenerIndex()
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
			command := a.commandsByName[fullName]
			a.RUnlock()

			lockDiff := time.Since(lockStart)
			metrics.ObserveLockWait("app_read", lockDiff)

			if lockDiff > 100*time.Millisecond {
				slog.Warn(
					"Locking app took too long",
					slog.String("app_id", appID),
					slog.String("lock_duration", lockDiff.String()),
				)
			}

			if command != nil {
				go command.HandleEvent(appID, session, event)
			}
		case *discord.ButtonInteraction:
			customID := string(d.CustomID)
			resumePointID, _, isResume := message.DecodeCustomIDMessageComponentResumePoint(customID)
			if isResume {
				a.resumeFlow(resumePointID, session, event)
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
		case *discord.StringSelectInteraction:
			// Select menus use the same resume-point system as buttons.
			// The CustomID on the select menu encodes the resume point;
			// the selected option's Value encodes the option's numeric ID.
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

			// Not a resume point — route via MessageInstance (message template)
			messageID := e.Message.ID.String()
			messageInstance, err := a.env.MessageInstanceStore.MessageInstanceByDiscordMessageID(context.TODO(), messageID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return
				}
				slog.With("error", err).Error("failed to get message instance for select menu")
				return
			}

			instance, err := NewMessageInstance(
				a.id,
				messageInstance,
				a.env,
			)
			if err != nil {
				slog.With("error", err).Error("failed to create message instance for select menu")
				return
			}

			go instance.HandleEvent(appID, session, event)
		case *discord.ModalInteraction:
			customID := string(d.CustomID)
			resumePointID, ok := message.DecodeCustomIDModalResumePoint(customID)
			if !ok {
				return
			}

			a.resumeFlow(resumePointID, session, event)
		}
	default:
		eventType := model.EventTypeFromDiscordEventType(e.EventType())

		// The index is replaced rather than mutated on rebuild, so this slice
		// stays valid after the lock is released.
		a.RLock()
		listeners := a.listenersByType[eventType]
		a.RUnlock()

		for _, listener := range listeners {
			go listener.HandleEvent(appID, session, event)
		}
	}
}

// resumeFlow loads a resume point and dispatches it back into the flow that
// created it.
//
// A resume point is owned by whatever ran the flow: a command, an event
// listener, or a message instance. Every owner has to be handled here — an
// unhandled one means the interaction never gets a response, which Discord
// shows to the user as "This interaction failed".
func (a *App) resumeFlow(
	resumePointID string,
	session *state.State,
	event gateway.Event,
) {
	resumePoint, err := a.env.ResumePointStore.ResumePoint(context.TODO(), resumePointID)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			slog.Error(
				"Failed to get resume point",
				slog.String("resume_point_id", resumePointID),
				slog.String("error", err.Error()),
			)
		}
		return
	}

	targetFlow := a.resumeFlowTarget(resumePoint)
	if targetFlow == nil {
		return
	}

	node := targetFlow.FindChildWithID(resumePoint.FlowNodeID, true)
	if node == nil {
		slog.Error(
			"Failed to find node in flow",
			slog.String("resume_point_id", resumePoint.ID),
			slog.String("flow_node_id", resumePoint.FlowNodeID),
		)
		return
	}

	go a.env.executeFlowEvent(
		context.Background(),
		a.id,
		node,
		session,
		event,
		entityLinksFromResumePoint(resumePoint),
		&resumePoint.FlowState,
	)
}

// entityLinksFromResumePoint recovers the links the flow was running with when
// it suspended. CreateResumePoint stores them verbatim, so attribution of logs
// and usage survives the suspend rather than being guessed on the way back in.
func entityLinksFromResumePoint(resumePoint *model.ResumePoint) entityLinks {
	return entityLinks{
		CommandID:         resumePoint.CommandID,
		EventListenerID:   resumePoint.EventListenerID,
		MessageID:         resumePoint.MessageID,
		MessageInstanceID: resumePoint.MessageInstanceID,
		FlowSourceID:      resumePoint.FlowSourceID,
	}
}

// resumeFlowTarget resolves which compiled flow a resume point belongs to, or
// nil if its owner is gone.
func (a *App) resumeFlowTarget(resumePoint *model.ResumePoint) *flow.CompiledFlowNode {
	switch {
	case resumePoint.CommandID.Valid:
		a.RLock()
		command, ok := a.commands[resumePoint.CommandID.String]
		a.RUnlock()
		if !ok {
			return nil
		}

		return command.flow
	case resumePoint.EventListenerID.Valid:
		a.RLock()
		listener, ok := a.listeners[resumePoint.EventListenerID.String]
		a.RUnlock()
		if !ok {
			return nil
		}

		return listener.flow
	case resumePoint.MessageInstanceID.Valid:
		messageInstance, err := a.env.MessageInstanceStore.MessageInstance(
			context.TODO(),
			resumePoint.MessageID.String,
			uint64(resumePoint.MessageInstanceID.Int64),
		)
		if err != nil {
			if !errors.Is(err, store.ErrNotFound) {
				slog.Error(
					"Failed to get message instance from resume point",
					slog.String("resume_point_id", resumePoint.ID),
					slog.String("message_id", resumePoint.MessageID.String),
					slog.Int64("message_instance_id", resumePoint.MessageInstanceID.Int64),
					slog.String("error", err.Error()),
				)
			}
			return nil
		}

		instance, err := NewMessageInstance(a.id, messageInstance, a.env)
		if err != nil {
			slog.Error(
				"Failed to create message instance",
				slog.String("resume_point_id", resumePoint.ID),
				slog.String("message_id", resumePoint.MessageID.String),
				slog.Int64("message_instance_id", resumePoint.MessageInstanceID.Int64),
				slog.String("error", err.Error()),
			)
			return nil
		}

		return instance.flows[resumePoint.FlowSourceID.String]
	default:
		slog.Error(
			"Resume point has no owning command, event listener or message instance",
			slog.String("resume_point_id", resumePoint.ID),
		)
		return nil
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
