package builder

import (
	"fmt"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type Command struct {
	cmd  *model.Command
	flow *flow.CompiledFlowNode
	env  Env
}

func NewCommand(
	cmd *model.Command,
	env Env,
) (*Command, error) {
	flow, err := flow.CompileCommand(cmd.FlowSource)
	if err != nil {
		slog.Error(
			"Failed to compile command flow",
			slog.String("app_id", cmd.AppID),
			slog.String("command_id", cmd.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to compile command flow: %w", err)
	}

	return &Command{
		cmd:  cmd,
		flow: flow,
		env:  env,
	}, nil
}

func (c *Command) HandleEvent(appID string, session *state.State, event gateway.Event) {
	links := entityLinks{
		CommandID: null.NewString(c.cmd.ID, true),
	}

	c.env.executeFlowEvent(
		c.cmd.AppID,
		c.flow,
		session,
		event,
		links,
		nil,
		false,
	)
}

type EventListener struct {
	listener *model.EventListener
	flow     *flow.CompiledFlowNode
	env      Env
}

func NewEventListener(
	listener *model.EventListener,
	env Env,
) (*EventListener, error) {
	flow, err := flow.CompileEventListener(listener.FlowSource)
	if err != nil {
		slog.Error(
			"Failed to compile event listener flow",
			slog.String("app_id", listener.AppID),
			slog.String("event_listener_id", listener.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to compile event listener flow: %w", err)
	}

	return &EventListener{
		listener: listener,
		flow:     flow,
		env:      env,
	}, nil
}

func (l *EventListener) HandleEvent(appID string, session *state.State, event gateway.Event) {
	links := entityLinks{
		EventListenerID: null.NewString(l.listener.ID, true),
	}

	// TODO: check listener specific filters as well
	if !l.shouldHandleEvent(event) {
		return
	}

	l.env.executeFlowEvent(
		l.listener.AppID,
		l.flow,
		session,
		event,
		links,
		nil,
		false,
	)
}

func (l *EventListener) shouldHandleEvent(e ws.Event) bool {
	switch d := e.(type) {
	case *gateway.MessageCreateEvent:
		// TODO?: It would be better if we check if the author is specifically the current app
		return !d.Author.Bot
	case *gateway.MessageUpdateEvent:
		return !d.Author.Bot
	case *gateway.MessageDeleteEvent:
		return true
	case *gateway.GuildMemberAddEvent:
		return true
	case *gateway.GuildMemberRemoveEvent:
		return true
	}

	return false
}

type MessageInstance struct {
	appID string
	msg   *model.MessageInstance
	flows map[string]*flow.CompiledFlowNode
	env   Env
}

func NewMessageInstance(
	appID string,
	msg *model.MessageInstance,
	env Env,
) (*MessageInstance, error) {
	flows := make(map[string]*flow.CompiledFlowNode, len(msg.FlowSources))

	for id, flowSource := range msg.FlowSources {
		flow, err := flow.CompileComponentButton(flowSource)
		if err != nil {
			slog.Error(
				"Failed to compile component button flow",
				slog.String("app_id", appID),
				slog.String("message_id", msg.MessageID),
				slog.String("error", err.Error()),
			)
			continue
		}

		flows[id] = flow
	}

	return &MessageInstance{
		appID: appID,
		msg:   msg,
		flows: flows,
		env:   env,
	}, nil
}

func (m *MessageInstance) HandleEvent(appID string, session *state.State, event gateway.Event) {
	i, ok := event.(*gateway.InteractionCreateEvent)
	if !ok {
		return
	}

	d, ok := i.InteractionEvent.Data.(*discord.ButtonInteraction)
	if !ok {
		return
	}

	flowSourceID := string(d.CustomID)

	links := entityLinks{
		MessageID:         null.NewString(m.msg.MessageID, true),
		MessageInstanceID: null.NewInt(int64(m.msg.ID), true),
		FlowSourceID:      null.NewString(flowSourceID, true),
	}

	targetFlow, ok := m.flows[flowSourceID]
	if !ok {
		return
	}

	m.env.executeFlowEvent(
		m.appID,
		targetFlow,
		session,
		event,
		links,
		nil,
		false,
	)
}
