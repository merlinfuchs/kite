package engine

import (
	"context"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
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
	flows := make(map[string]*flow.CompiledFlowNode, len(msg.FlowSources))

	flowSources := withTemplateParameters(msg.FlowSources, msg.TemplateFlowSources)
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

// withTemplateParameters returns the flows of a sent message with the values
// the parameters have in its message template now. Changing a value in the
// template then also changes messages that were sent before, while their
// blocks stay as they were when the message was sent.
func withTemplateParameters(flowSources, template map[string]flow.FlowData) map[string]flow.FlowData {
	res := make(map[string]flow.FlowData, len(flowSources))
	for id, flowSource := range flowSources {
		if current, ok := template[id]; ok {
			flowSource, _ = flowSource.WithParameterValues(current.ParameterValues())
		}
		res[id] = flowSource
	}
	return res
}
