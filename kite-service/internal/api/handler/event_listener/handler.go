package eventlistener

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
)

type EventListenerHandler struct {
	eventListenerStore store.EventListenerStore
	webhookRunner      WebhookRunner
	webhookLimiter     *webhookLimiter
	// How long a webhook request waits for its flow to respond.
	webhookResponseTimeout time.Duration
}

func NewEventListenerHandler(eventListenerStore store.EventListenerStore, webhookRunner WebhookRunner) *EventListenerHandler {
	return &EventListenerHandler{
		eventListenerStore:     eventListenerStore,
		webhookRunner:          webhookRunner,
		webhookLimiter:         newWebhookLimiter(),
		webhookResponseTimeout: webhook.ResponseTimeout,
	}
}

func (h *EventListenerHandler) HandleEventListenerList(c *handler.Context) (*wire.EventListenerListResponse, error) {
	eventListeners, err := h.eventListenerStore.EventListenersByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event listeners: %w", err)
	}

	res := make([]*wire.EventListener, len(eventListeners))
	for i, eventListener := range eventListeners {
		res[i] = wire.EventListenerToWire(eventListener)
	}

	return &res, nil
}

func (h *EventListenerHandler) HandleEventListenerGet(c *handler.Context) (*wire.EventListenerGetResponse, error) {
	return wire.EventListenerToWire(c.EventListener), nil
}

func (h *EventListenerHandler) HandleEventListenerCreate(c *handler.Context, req wire.EventListenerCreateRequest) (*wire.EventListenerCreateResponse, error) {
	source := model.EventSource(req.Source)
	eventFlow, err := compileEventListener(c, source, req.FlowSource)
	if err != nil {
		return nil, err
	}

	if err := h.checkEventListenerLimit(c, map[model.EventSource]int{source: 1}); err != nil {
		return nil, err
	}

	eventListener, err := h.eventListenerStore.CreateEventListener(c.Context(), &model.EventListener{
		ID:            util.UniqueID(),
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		Source:        source,
		Type:          model.EventListenerType(eventFlow.EventListenerType()),
		Description:   eventFlow.EventDescription(),
		// TODO: Filter:        eventFlow.EventListenerFilter(),
		FlowSource:    req.FlowSource,
		Enabled:       req.Enabled,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		WebhookSecret: newWebhookSecret(source),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event listener: %w", err)
	}

	return wire.EventListenerToWire(eventListener), nil
}

func (h *EventListenerHandler) HandleEventListenersImport(c *handler.Context, req wire.EventListenersImportRequest) (*wire.EventListenersImportResponse, error) {
	eventListeners, err := h.PrepareImport(c, req.EventListeners)
	if err != nil {
		return nil, err
	}

	if err := h.CreateEventListeners(c, eventListeners); err != nil {
		return nil, err
	}

	res := make([]*wire.EventListener, len(eventListeners))
	for i, eventListener := range eventListeners {
		res[i] = wire.EventListenerToWire(eventListener)
	}

	return &res, nil
}

// PrepareImport checks that the event listeners can be added to the app and
// returns them with their IDs set, without saving them.
func (h *EventListenerHandler) PrepareImport(c *handler.Context, reqs []wire.EventListenerCreateRequest) ([]*model.EventListener, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	eventListeners := make([]*model.EventListener, len(reqs))
	added := make(map[model.EventSource]int)

	for i, listener := range reqs {
		source := model.EventSource(listener.Source)
		eventFlow, err := compileEventListener(c, source, listener.FlowSource)
		if err != nil {
			return nil, err
		}
		added[source]++

		eventListeners[i] = &model.EventListener{
			ID:            util.UniqueID(),
			AppID:         c.App.ID,
			CreatorUserID: c.Session.UserID,
			Source:        source,
			Type:          model.EventListenerType(eventFlow.EventListenerType()),
			Description:   eventFlow.EventDescription(),
			// TODO: Filter:        eventFlow.EventListenerFilter(),
			FlowSource:    listener.FlowSource,
			Enabled:       listener.Enabled,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			WebhookSecret: newWebhookSecret(source),
		}
	}

	if err := h.checkEventListenerLimit(c, added); err != nil {
		return nil, err
	}

	return eventListeners, nil
}

// CreateEventListeners saves the event listeners from PrepareImport.
func (h *EventListenerHandler) CreateEventListeners(c *handler.Context, eventListeners []*model.EventListener) error {
	for i, eventListener := range eventListeners {
		created, err := h.eventListenerStore.CreateEventListener(c.Context(), eventListener)
		if err != nil {
			return fmt.Errorf("failed to create event listener: %w", err)
		}
		eventListeners[i] = created
	}

	return nil
}

func (h *EventListenerHandler) HandleEventListenerUpdate(c *handler.Context, req wire.EventListenerUpdateRequest) (*wire.EventListenerUpdateResponse, error) {
	updated, err := h.PrepareUpdate(c, c.EventListener, req.FlowSource)
	if err != nil {
		return nil, err
	}

	eventListener, err := h.eventListenerStore.UpdateEventListener(c.Context(), updated)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_event_listener", "Event listener not found")
		}
		return nil, fmt.Errorf("failed to update event listener: %w", err)
	}

	return wire.EventListenerToWire(eventListener), nil
}

// PrepareUpdate checks that the listener can be changed to the flow and
// returns it changed, without saving it.
func (h *EventListenerHandler) PrepareUpdate(c *handler.Context, listener *model.EventListener, flowSource flow.FlowData) (*model.EventListener, error) {
	// The source is fixed at creation because it decides which limit the
	// listener counts against.
	eventFlow, err := compileEventListener(c, listener.Source, flowSource)
	if err != nil {
		return nil, err
	}

	updated := *listener
	updated.Type = model.EventListenerType(eventFlow.EventListenerType())
	updated.Description = eventFlow.EventDescription()
	// TODO: updated.Filter = eventFlow.EventListenerFilter()
	updated.FlowSource = flowSource
	updated.UpdatedAt = time.Now().UTC()
	return &updated, nil
}

func (h *EventListenerHandler) HandleEventListenerUpdateEnabled(c *handler.Context, req wire.EventListenerUpdateEnabledRequest) (*wire.EventListenerUpdateEnabledResponse, error) {
	eventListener, err := h.eventListenerStore.UpdateEventListener(c.Context(), &model.EventListener{
		ID:          c.EventListener.ID,
		Type:        c.EventListener.Type,
		Description: c.EventListener.Description,
		FlowSource:  c.EventListener.FlowSource,
		Enabled:     req.Enabled,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_event_listener", "Event listener not found")
		}
		return nil, fmt.Errorf("failed to update event listener: %w", err)
	}

	return wire.EventListenerToWire(eventListener), nil
}

func (h *EventListenerHandler) HandleEventListenerDelete(c *handler.Context) (*wire.EventListenerDeleteResponse, error) {
	err := h.eventListenerStore.DeleteEventListener(c.Context(), c.EventListener.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_event_listener", "Event listener not found")
		}
		return nil, fmt.Errorf("failed to delete event listener: %w", err)
	}

	return &wire.EventListenerDeleteResponse{}, nil
}

// compileEventListener compiles the flow and checks that it matches the
// requested source and that its event filters fit its event. Scheduled
// listeners are also checked against the app's minimum schedule interval.
func compileEventListener(c *handler.Context, source model.EventSource, flowSource flow.FlowData) (*flow.CompiledFlowNode, error) {
	eventFlow, err := flow.CompileEventListener(flowSource)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_flow", err.Error())
	}

	if typeSource := model.EventSourceForType(model.EventListenerType(eventFlow.EventListenerType())); typeSource != source {
		return nil, handler.ErrBadRequest(
			"invalid_source",
			fmt.Sprintf("event type %s belongs to source %s, not %s", eventFlow.EventListenerType(), typeSource, source),
		)
	}

	// Checked on save only, so listeners saved before this can still be
	// loaded and get the error the next time they're edited.
	if err := flow.ValidateEventFilters(eventFlow); err != nil {
		return nil, handler.ErrBadRequest("invalid_event_filter", err.Error())
	}

	if source == model.EventSourceSchedule {
		sched, err := schedule.Parse(eventFlow.EventScheduleCron())
		if err != nil {
			return nil, handler.ErrBadRequest("invalid_schedule", err.Error())
		}

		minInterval := c.Features.MinScheduleInterval()
		gap := sched.MinGap(time.Now().UTC())
		if gap == 0 {
			return nil, handler.ErrBadRequest("invalid_schedule", "schedule never runs")
		}
		if gap < minInterval {
			return nil, handler.ErrBadRequest(
				"schedule_too_frequent",
				fmt.Sprintf("schedule runs every %s, but your plan allows at most one run every %s", gap, minInterval),
			)
		}
	}

	return eventFlow, nil
}

// checkEventListenerLimit checks that the app can have added more listeners
// of each source. Each source has its own limit.
func (h *EventListenerHandler) checkEventListenerLimit(c *handler.Context, added map[model.EventSource]int) error {
	for source, count := range added {
		// Named like the pages they're on in the dashboard, so the limit that
		// was reached is clear.
		limit, name := c.Features.MaxEventListeners, "Discord events"
		switch source {
		case model.EventSourceSchedule:
			limit, name = c.Features.MaxScheduledEventListeners, "scheduled tasks"
		case model.EventSourceWebhook:
			limit, name = c.Features.MaxWebhookEventListeners, "webhooks"
		}
		if limit == 0 {
			continue
		}

		existing, err := h.eventListenerStore.CountEventListenersByAppAndSource(c.Context(), c.App.ID, source)
		if err != nil {
			return fmt.Errorf("failed to count event listeners: %w", err)
		}

		if existing+count > limit {
			return handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of %s (%d) reached", name, limit))
		}
	}

	return nil
}
