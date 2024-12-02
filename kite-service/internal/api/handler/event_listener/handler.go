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
)

type EventListenerHandler struct {
	eventListenerStore      store.EventListenerStore
	maxEventListenersPerApp int
}

func NewEventListenerHandler(eventListenerStore store.EventListenerStore, maxEventListenersPerApp int) *EventListenerHandler {
	return &EventListenerHandler{
		eventListenerStore:      eventListenerStore,
		maxEventListenersPerApp: maxEventListenersPerApp,
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
	if h.maxEventListenersPerApp != 0 {
		eventListenerCount, err := h.eventListenerStore.CountEventListenersByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count event listeners: %w", err)
		}

		if eventListenerCount >= h.maxEventListenersPerApp {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of event listeners (%d) reached", h.maxEventListenersPerApp))
		}
	}

	eventFlow, err := flow.CompileEventListener(req.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile event listener: %w", err)
	}

	eventListener, err := h.eventListenerStore.CreateEventListener(c.Context(), &model.EventListener{
		ID:            util.UniqueID(),
		Name:          eventFlow.EventListenerName(),
		Description:   eventFlow.EventListenerDescription(),
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		Integration:   model.IntegrationType(req.Integration),
		Type:          model.EventListenerType(eventFlow.EventListenerType()),
		// TODO: Filter:        eventFlow.EventListenerFilter(),
		FlowSource: req.FlowSource,
		Enabled:    req.Enabled,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event listener: %w", err)
	}

	return wire.EventListenerToWire(eventListener), nil
}

func (h *EventListenerHandler) HandleEventListenerUpdate(c *handler.Context, req wire.EventListenerUpdateRequest) (*wire.EventListenerUpdateResponse, error) {
	eventFlow, err := flow.CompileEventListener(req.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile event listener: %w", err)
	}

	eventListener, err := h.eventListenerStore.UpdateEventListener(c.Context(), &model.EventListener{
		ID:          c.EventListener.ID,
		Name:        eventFlow.EventListenerName(),
		Description: eventFlow.EventListenerDescription(),
		Type:        model.EventListenerType(eventFlow.EventListenerType()),
		// TODO: Filter:      eventFlow.EventListenerFilter(),
		FlowSource: req.FlowSource,
		Enabled:    req.Enabled,
		UpdatedAt:  time.Now().UTC(),
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
