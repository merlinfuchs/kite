package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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
	scheduled       []*EventListener
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

// rebuildListenerIndex regenerates the event type lookup of Discord listeners
// and the list of scheduled listeners from a.listeners. Callers must hold the
// write lock.
func (a *App) rebuildListenerIndex() {
	index := make(map[model.EventListenerType][]*EventListener, len(a.listeners))
	var scheduled []*EventListener
	for _, listener := range a.listeners {
		switch listener.listener.Source {
		case model.EventSourceDiscord:
			index[listener.listener.Type] = append(index[listener.listener.Type], listener)
		case model.EventSourceSchedule:
			scheduled = append(scheduled, listener)
		}
	}
	a.listenersByType = index
	a.scheduled = scheduled
}

func (a *App) scheduledEventListeners() []*EventListener {
	a.RLock()
	defer a.RUnlock()
	return a.scheduled
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

func (a *App) RemovePluginInstance(pluginInstanceID string) {
	a.Lock()
	defer a.Unlock()

	pluginInstance, ok := a.pluginInstances[pluginInstanceID]
	if !ok {
		return
	}

	if err := pluginInstance.Close(); err != nil {
		slog.With("error", err).Error("failed to close plugin instance")
	}
	delete(a.pluginInstances, pluginInstanceID)
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

func (a *App) RemoveCommand(commandID string) {
	a.Lock()
	defer a.Unlock()

	if _, ok := a.commands[commandID]; ok {
		delete(a.commands, commandID)
		a.rebuildCommandIndex()
	}
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

	if old, ok := a.listeners[listenerID]; ok && old.schedule != nil && listener.schedule != nil {
		listener.schedule.takeOver(old.schedule)
	}

	a.listeners[listenerID] = listener
	a.rebuildListenerIndex()
}

func (a *App) RemoveEventListener(listenerID string) {
	a.Lock()
	defer a.Unlock()

	if _, ok := a.listeners[listenerID]; ok {
		delete(a.listeners, listenerID)
		a.rebuildListenerIndex()
	}
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
		case discord.ComponentInteraction:
			customID := string(d.ID())
			resumePointID, _, isResume := message.DecodeCustomIDMessageComponentResumePoint(customID)
			if isResume {
				a.resumeFlowOrRespondExpired(resumePointID, session, e)
				return
			}

			messageID := e.Message.ID.String()
			messageInstnace, err := a.env.MessageInstanceStore.MessageInstanceByDiscordMessageID(context.TODO(), a.id, messageID)
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

			a.touchMessageInstance(messageInstnace)
			go instance.HandleEvent(appID, session, event)
		case *discord.ModalInteraction:
			customID := string(d.CustomID)
			resumePointID, ok := message.DecodeCustomIDModalResumePoint(customID)
			if !ok {
				return
			}

			a.resumeFlowOrRespondExpired(resumePointID, session, e)
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
// created it. It returns false if the resume point doesn't exist, usually
// because it expired.
//
// A resume point is owned by whatever ran the flow: a command, an event
// listener, or a message instance. Every owner has to be handled here — an
// unhandled one means the interaction never gets a response, which Discord
// shows to the user as "This interaction failed".
func (a *App) resumeFlow(
	resumePointID string,
	session *state.State,
	event gateway.Event,
) bool {
	// The ID comes from a user-controlled custom_id, so the lookup must be scoped to the app.
	resumePoint, err := a.env.ResumePointStore.ResumePoint(context.TODO(), a.id, resumePointID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return false
		}

		slog.Error(
			"Failed to get resume point",
			slog.String("resume_point_id", resumePointID),
			slog.String("error", err.Error()),
		)
		return true
	}

	node := a.resumeNode(resumePoint)
	if node == nil {
		return true
	}

	a.touchResumePoint(resumePoint)

	// A click or submit starts a new execution, so its durable sleeps count
	// from zero. Menus whose buttons wait before sending the next menu would
	// otherwise run into the limit after a few clicks.
	resumePoint.FlowState.DurableSleeps = 0

	go a.env.executeFlowEvent(
		context.Background(),
		a.id,
		node,
		session,
		event,
		entityLinksFromResumePoint(resumePoint),
		&resumePoint.FlowState,
	)
	return true
}

// resumeNode finds the node a resume point continues from, or nil if its flow
// or the node itself is gone.
func (a *App) resumeNode(resumePoint *model.ResumePoint) *flow.CompiledFlowNode {
	targetFlow := a.resumeFlowTarget(resumePoint)
	if targetFlow == nil {
		return nil
	}

	node := targetFlow.FindChildWithID(resumePoint.FlowNodeID, true)
	if node == nil {
		slog.Error(
			"Failed to find node in flow",
			slog.String("resume_point_id", resumePoint.ID),
			slog.String("flow_node_id", resumePoint.FlowNodeID),
		)
	}
	return node
}

// resumeFlowAfterSleep continues a flow whose durable sleep is over. The timer
// is leased, and only deleted once everything needed to resume was found, so
// a timer that can't resume yet, e.g. because the engine is still loading,
// is retried when the lease is over.
func (a *App) resumeFlowAfterSleep(resumePoint *model.ResumePoint, session *state.State) {
	// Deleting the owner clears its link, so the timer can never resume.
	if !resumePoint.CommandID.Valid && !resumePoint.EventListenerID.Valid && !resumePoint.MessageInstanceID.Valid {
		a.dropTimer(resumePoint, "its command, event listener or message was deleted")
		return
	}

	// Not loaded yet or disabled, retried after the lease.
	targetFlow := a.resumeFlowTarget(resumePoint)
	if targetFlow == nil {
		return
	}

	node := targetFlow.FindChildWithID(resumePoint.FlowNodeID, true)
	if node == nil {
		a.dropTimer(resumePoint, "its Wait block was deleted")
		return
	}

	event, state, err := a.timerResumeEvent(resumePoint)
	if err != nil {
		a.dropTimer(resumePoint, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	deleted, err := a.env.ResumePointStore.DeleteTimerResumePoint(ctx, a.id, resumePoint.ID)
	if err != nil {
		slog.Error(
			"Failed to delete timer resume point",
			slog.String("resume_point_id", resumePoint.ID),
			slog.String("error", err.Error()),
		)
		return
	}
	if !deleted {
		// Resumed by someone else after the lease ran out.
		return
	}

	a.env.executeFlowAfterSleep(
		context.Background(),
		a.id,
		node,
		session,
		event,
		entityLinksFromResumePoint(resumePoint),
		&state,
	)
}

// dropTimer deletes a timer that can never resume, so it doesn't count
// against the app's limit until it expires.
func (a *App) dropTimer(resumePoint *model.ResumePoint, reason string) {
	slog.Warn(
		"Dropping timer resume point that can't resume",
		slog.String("app_id", a.id),
		slog.String("resume_point_id", resumePoint.ID),
		slog.String("reason", reason),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := a.env.ResumePointStore.DeleteTimerResumePoint(ctx, a.id, resumePoint.ID); err != nil {
		slog.Error(
			"Failed to delete timer resume point",
			slog.String("resume_point_id", resumePoint.ID),
			slog.String("error", err.Error()),
		)
	}
}

// timerResumeEvent restores the interaction or event a flow suspended with in
// a durable sleep.
func (a *App) timerResumeEvent(resumePoint *model.ResumePoint) (gateway.Event, flow.FlowContextState, error) {
	state := resumePoint.FlowState
	trigger := state.ResumeTrigger
	if trigger == nil {
		return nil, state, errors.New("resume point has no trigger")
	}
	state.ResumeTrigger = nil

	if trigger.Interaction == nil {
		return trigger.Event, state, nil
	}

	interaction := *trigger.Interaction
	if resumePoint.InteractionToken.Valid {
		token, err := a.env.TokenCrypt.DecryptString(resumePoint.InteractionToken.String)
		if err != nil {
			return nil, state, fmt.Errorf("failed to decrypt interaction token: %w", err)
		}
		interaction.Token = token
	}
	return &gateway.InteractionCreateEvent{InteractionEvent: interaction}, state, nil
}

// resumeFlowOrRespondExpired tells the user why nothing happened when the resume
// point is gone, instead of leaving them with Discord's generic "This
// interaction failed".
func (a *App) resumeFlowOrRespondExpired(resumePointID string, session *state.State, e *gateway.InteractionCreateEvent) {
	if a.resumeFlow(resumePointID, session, e) {
		return
	}

	id, token := e.ID, e.Token
	go func() {
		err := session.RespondInteraction(id, token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString("This has expired. Run the command again or ask an admin to send the message again."),
				Flags:   discord.EphemeralMessage,
			},
		})
		if err != nil {
			slog.Error(
				"Failed to respond to expired resume point",
				slog.String("app_id", a.id),
				slog.String("error", err.Error()),
			)
		}
	}()
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
			a.id,
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

		// Otherwise the instance could expire while its resume points are still in use.
		a.touchMessageInstance(messageInstance)

		return instance.flows[resumePoint.FlowSourceID.String]
	default:
		slog.Error(
			"Resume point has no owning command, event listener or message instance",
			slog.String("resume_point_id", resumePoint.ID),
		)
		return nil
	}
}

// touchInterval limits last_used_at writes so busy buttons don't write on every
// click. Unused resume points and message instances are deleted by the usage
// manager.
const touchInterval = time.Hour

func (a *App) touchResumePoint(resumePoint *model.ResumePoint) {
	id := resumePoint.ID
	a.touch(resumePoint.LastUsedAt, func(ctx context.Context, now time.Time) error {
		return a.env.ResumePointStore.TouchResumePoint(ctx, a.id, id, now)
	})
}

func (a *App) touchMessageInstance(instance *model.MessageInstance) {
	id := instance.ID
	a.touch(instance.LastUsedAt, func(ctx context.Context, now time.Time) error {
		return a.env.MessageInstanceStore.TouchMessageInstance(ctx, a.id, id, now)
	})
}

// touch runs in the background so it doesn't delay the interaction response.
func (a *App) touch(lastUsedAt time.Time, update func(ctx context.Context, now time.Time) error) {
	now := time.Now().UTC()
	if now.Sub(lastUsedAt) < touchInterval {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := update(ctx, now); err != nil {
			slog.Error(
				"Failed to update last used time",
				slog.String("app_id", a.id),
				slog.String("error", err.Error()),
			)
		}
	}()
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
