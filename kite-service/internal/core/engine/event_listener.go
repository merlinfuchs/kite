package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"gopkg.in/guregu/null.v4"
)

type EventListener struct {
	listener *model.EventListener
	flow     *flow.CompiledFlowNode
	env      Env

	// schedule is only set for scheduled listeners.
	schedule *listenerSchedule
}

// NewEventListener compiles a listener. catchUp lets a scheduled listener run
// the latest occurrence it missed while the engine was down.
func NewEventListener(
	listener *model.EventListener,
	env Env,
	catchUp bool,
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

	res := &EventListener{
		listener: listener,
		flow:     flow,
		env:      env,
	}

	if listener.Source == model.EventSourceSchedule {
		res.schedule, err = newListenerSchedule(flow.EventScheduleCron(), listener.LastRunAt, catchUp)
		if err != nil {
			slog.Error(
				"Failed to parse event listener schedule",
				slog.String("app_id", listener.AppID),
				slog.String("event_listener_id", listener.ID),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to parse event listener schedule: %w", err)
		}
	}

	return res, nil
}

func (l *EventListener) HandleScheduledRun(session *state.State, occurrence time.Time) {
	l.env.executeFlowEvent(
		context.Background(),
		l.listener.AppID,
		l.flow,
		session,
		&schedule.Event{Time: occurrence},
		entityLinks{
			EventListenerID: null.NewString(l.listener.ID, true),
		},
		nil,
	)
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
		context.Background(),
		l.listener.AppID,
		l.flow,
		session,
		event,
		links,
		nil,
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
