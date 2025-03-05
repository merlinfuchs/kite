package builder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"gopkg.in/guregu/null.v4"
)

type BuilderPluginInstance struct {
	sync.RWMutex

	appID  string
	config plugin.ConfigValues
	env    Env

	lastUpdate time.Time
	commands   map[string]*Command
	listeners  map[string]*EventListener

	commandData []plugin.Command
}

func newBuilderPluginInstance(
	appID string,
	config plugin.ConfigValues,
	env Env,
) *BuilderPluginInstance {
	return &BuilderPluginInstance{
		appID:  appID,
		config: config,
		env:    env,

		commands:  make(map[string]*Command),
		listeners: make(map[string]*EventListener),
	}
}

func (p *BuilderPluginInstance) Events() []plugin.Event {
	return []plugin.Event{
		{
			ID: "builder_command_created",
		},
	}
}

func (p *BuilderPluginInstance) Commands() []plugin.Command {
	return p.commandData
}

func (p *BuilderPluginInstance) Update(c plugin.Context, config plugin.ConfigValues) error {
	p.Lock()
	p.config = config
	lastUpdate := p.lastUpdate
	p.lastUpdate = time.Now().UTC()
	p.Unlock()

	err := p.populateCommands(c, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to populate commands: %w", err)
	}

	err = p.populateEventListeners(c, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to populate event listeners: %w", err)
	}

	err = p.updateCommandData()
	if err != nil {
		return fmt.Errorf("failed to update command data: %w", err)
	}

	return nil
}

func (p *BuilderPluginInstance) updateCommandData() error {
	p.RLock()

	var lastUpdatedAt time.Time

	commandNames := make([]string, 0, len(p.commands))
	commands := make([]plugin.Command, 0, len(p.commands))
	for _, command := range p.commands {
		cmd := command.cmd

		if cmd.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = cmd.UpdatedAt
		}

		node, err := flow.CompileCommand(cmd.FlowSource)
		if err != nil {
			return fmt.Errorf("failed to compile command flow: %w", err)
		}

		data := node.CommandData()
		commands = append(commands, plugin.Command{
			ID: cmd.ID,
			Data: api.CreateCommandData{
				Name:                     data.Name,
				Description:              data.Description,
				Options:                  data.Options,
				DefaultMemberPermissions: data.DefaultMemberPermissions,
				Contexts:                 node.CommandContexts(),
				IntegrationTypes:         node.CommandIntegrations(),
			},
		})
		commandNames = append(commandNames, node.CommandName())
	}

	p.RUnlock()

	if err := validateCommandNames(commandNames); err != nil {
		return fmt.Errorf("failed to validate command names: %w", err)
	}

	p.Lock()
	p.commandData = commands
	p.Unlock()

	return nil
}

func (p *BuilderPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	p.RLock()
	defer p.RUnlock()

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		switch d := e.Data.(type) {
		case *discord.CommandInteraction:
			fullName := getFullCommandName(d)
			for _, command := range p.commands {
				if command.cmd.Name == fullName {
					go command.HandleEvent(p.appID, c.Discord(), event)
				}
			}
		case *discord.ButtonInteraction:
			messageID := e.Message.ID.String()
			messageInstnace, err := p.env.MessageInstanceStore.MessageInstanceByDiscordMessageID(context.TODO(), messageID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return nil
				}

				slog.Error(
					"Failed to get message instance by discord message ID",
					slog.String("message_id", messageID),
					slog.String("error", err.Error()),
				)
				return fmt.Errorf("failed to get message instance by discord message ID: %w", err)
			}

			instance, err := NewMessageInstance(
				p.appID,
				messageInstnace,
				p.env,
			)
			if err != nil {
				slog.Error(
					"Failed to create message instance",
					slog.String("message_id", messageID),
					slog.String("error", err.Error()),
				)
				return fmt.Errorf("failed to create message instance: %w", err)
			}

			go instance.HandleEvent(p.appID, c.Discord(), event)
		case *discord.ModalInteraction:
			customID := string(d.CustomID)
			if !strings.HasPrefix(customID, "resume:") {
				return nil
			}

			resumePointID := customID[len("resume:"):]
			resumePoint, err := p.env.ResumePointStore.ResumePoint(context.TODO(), resumePointID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return nil
				}

				slog.Error(
					"Failed to get resume point",
					slog.String("resume_point_id", resumePointID),
					slog.String("error", err.Error()),
				)
				return fmt.Errorf("failed to get resume point: %w", err)
			}

			if resumePoint.CommandID.Valid {
				command, ok := p.commands[resumePoint.CommandID.String]
				if !ok {
					return nil
				}

				node := command.flow.FindChildWithID(resumePoint.FlowNodeID)

				p.env.executeFlowEvent(
					p.appID,
					node,
					c.Discord(),
					event,
					entityLinks{
						CommandID: null.NewString(command.cmd.ID, true),
					},
					&resumePoint.FlowState,
					true,
				)
			}

			if resumePoint.MessageInstanceID.Valid {
				messageInstance, err := p.env.MessageInstanceStore.MessageInstance(
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
						return fmt.Errorf("failed to get message instance from resume point: %w", err)
					}
					return nil
				}

				instance, err := NewMessageInstance(
					p.appID,
					messageInstance,
					p.env,
				)
				if err != nil {
					slog.Error(
						"Failed to create message instance",
						slog.String("resume_point_id", resumePointID),
						slog.String("message_id", resumePoint.MessageID.String),
						slog.Int64("message_instance_id", resumePoint.MessageInstanceID.Int64),
						slog.String("error", err.Error()),
					)
					return fmt.Errorf("failed to create message instance: %w", err)
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
					return fmt.Errorf("failed to get target flow from resume point: %w", err)
				}

				node := targetFlow.FindChildWithID(resumePoint.FlowNodeID)

				p.env.executeFlowEvent(
					p.appID,
					node,
					c.Discord(),
					event,
					entityLinks{},
					&resumePoint.FlowState,
					true,
				)
			}
		}
	default:
		eventType := model.EventTypeFromDiscordEventType(e.EventType())
		for _, listener := range p.listeners {
			if listener.listener.Source != model.EventSourceDiscord {
				continue
			}

			if listener.listener.Type != eventType {
				continue
			}

			listener.HandleEvent(p.appID, c.Discord(), event)
		}
	}

	return nil
}

func (m *BuilderPluginInstance) populateCommands(ctx context.Context, lastUpdate time.Time) error {
	commandIDs, err := m.env.CommandStore.EnabledCommandIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled command IDs: %w", err)
	}

	// TODO: only get commands that belong to this app
	commands, err := m.env.CommandStore.EnabledCommandsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	m.Lock()
	defer m.Unlock()

	for _, command := range commands {
		m.addCommand(command)
	}

	m.removeDanglingCommands(commandIDs)

	return nil
}

func (p *BuilderPluginInstance) populateEventListeners(ctx context.Context, lastUpdate time.Time) error {
	listenerIDs, err := p.env.EventListenerStore.EnabledEventListenerIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled event listener IDs: %w", err)
	}

	listeners, err := p.env.EventListenerStore.EnabledEventListenersUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get event listeners: %w", err)
	}

	p.Lock()
	defer p.Unlock()

	for _, listener := range listeners {
		p.addEventListener(listener)
	}

	p.removeDanglingEventListeners(listenerIDs)

	return nil
}

func (p *BuilderPluginInstance) addCommand(cmd *model.Command) bool {
	command, err := NewCommand(
		cmd,
		p.env,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create command")
		return false
	}

	p.Lock()
	defer p.Unlock()

	p.commands[cmd.ID] = command

	if !cmd.LastDeployedAt.Valid || cmd.LastDeployedAt.Time.Before(cmd.UpdatedAt) {
		return true
	}

	return false
}

func (p *BuilderPluginInstance) removeDanglingCommands(commandIDs []string) bool {
	commandIDMap := make(map[string]struct{}, len(commandIDs))
	for _, commandID := range commandIDs {
		commandIDMap[commandID] = struct{}{}
	}

	p.Lock()
	defer p.Unlock()

	changed := false
	for cmdID := range p.commands {
		if _, ok := commandIDMap[cmdID]; !ok {
			delete(p.commands, cmdID)
			changed = true
		}
	}

	return changed
}

func (p *BuilderPluginInstance) addEventListener(listener *model.EventListener) {
	eventListener, err := NewEventListener(
		listener,
		p.env,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create event listener")
		return
	}

	p.Lock()
	defer p.Unlock()

	p.listeners[listener.ID] = eventListener
}

func (p *BuilderPluginInstance) removeDanglingEventListeners(listenerIDs []string) {
	listenerIDMap := make(map[string]struct{}, len(listenerIDs))
	for _, listenerID := range listenerIDs {
		listenerIDMap[listenerID] = struct{}{}
	}

	p.Lock()
	defer p.Unlock()

	for listenerID := range p.listeners {
		if _, ok := listenerIDMap[listenerID]; !ok {
			delete(p.listeners, listenerID)
		}
	}
}
