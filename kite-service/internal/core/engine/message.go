package engine

import (
	"context"
	"errors"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

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

	d, ok := i.InteractionEvent.Data.(discord.ComponentInteraction)
	if !ok {
		return
	}

	flowSourceID := m.flowSourceID(i, d)

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
		context.Background(),
		m.appID,
		targetFlow,
		session,
		event,
		links,
		nil,
	)
}

// flowSourceID returns the flow a used component runs. Options of a string
// select have their own flows like buttons, the first picked option with one
// runs it. Otherwise the select menu's own flow runs, which menus made before
// options had flows use to branch on interaction.value.
func (m *MessageInstance) flowSourceID(i *gateway.InteractionCreateEvent, d discord.ComponentInteraction) string {
	flowSourceID := string(d.ID())

	picked, ok := d.(*discord.StringSelectInteraction)
	if !ok {
		return flowSourceID
	}

	// The instance only keeps the flows, the options are in the template.
	msg, err := m.env.MessageStore.Message(context.Background(), m.appID, m.msg.MessageID)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			slog.Error(
				"Failed to get message for select menu options",
				slog.String("app_id", m.appID),
				slog.String("message_id", m.msg.MessageID),
				slog.String("error", err.Error()),
			)
		}
		return flowSourceID
	}

	menu := msg.Data.FindComponent(func(c *message.ComponentData) bool {
		return c.FlowSourceID == flowSourceID && c.Type == message.ComponentTypeStringSelect
	})
	if menu == nil {
		return flowSourceID
	}

	sent := message.FindSentStringSelect(i.Message, d.ID())
	option := menu.PickedOption(picked.Values, sent, func(o *message.ComponentSelectOptionData) bool {
		_, ok := m.flows[o.FlowSourceID]
		return o.FlowSourceID != "" && ok
	})
	if option == nil {
		return flowSourceID
	}

	return option.FlowSourceID
}
