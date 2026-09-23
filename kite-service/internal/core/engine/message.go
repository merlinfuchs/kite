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
	flowSources := liveFlowSources(env.MessageStore, msg)
	flows := make(map[string]*flow.CompiledFlowNode, len(flowSources))

	for id, flowSource := range flowSources {
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

// liveFlowSources returns the instance's flows as they are in the template now,
// so edits to a template apply to messages that were already sent. The
// snapshot taken at send time is only used for components that were removed
// from the template since, or if the template can't be loaded.
func liveFlowSources(messageStore store.MessageStore, msg *model.MessageInstance) map[string]flow.FlowData {
	template, err := messageStore.Message(context.TODO(), msg.MessageID)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			slog.Error(
				"Failed to get message template for instance",
				slog.String("message_id", msg.MessageID),
				slog.String("error", err.Error()),
			)
		}
		return msg.FlowSources
	}

	flowSources := make(map[string]flow.FlowData, len(msg.FlowSources))
	for id, flowSource := range msg.FlowSources {
		if live, ok := template.FlowSources[id]; ok {
			flowSource = live
		}
		flowSources[id] = flowSource
	}
	return flowSources
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

	flowSourceID := string(d.ID())

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
