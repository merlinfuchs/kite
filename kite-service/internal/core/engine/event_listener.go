package engine

import (
	"fmt"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

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
