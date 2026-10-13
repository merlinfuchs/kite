package support

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type SupportHandler struct {
	appStore           store.AppStore
	logStore           store.LogStore
	commandStore       store.CommandStore
	eventListenerStore store.EventListenerStore
}

func NewSupportHandler(
	appStore store.AppStore,
	logStore store.LogStore,
	commandStore store.CommandStore,
	eventListenerStore store.EventListenerStore,
) *SupportHandler {
	return &SupportHandler{
		appStore:           appStore,
		logStore:           logStore,
		commandStore:       commandStore,
		eventListenerStore: eventListenerStore,
	}
}

type AccessResponse struct {
	Authorized bool `json:"authorized"`
}

type EventListener struct {
	ID            string                    `json:"id"`
	Source        string                    `json:"source"`
	Type          string                    `json:"type"`
	Description   string                    `json:"description"`
	Enabled       bool                      `json:"enabled"`
	AppID         string                    `json:"app_id"`
	ModuleID      null.String               `json:"module_id"`
	CreatorUserID string                    `json:"creator_user_id"`
	Filter        *wire.EventListenerFilter `json:"filter"`
	FlowSource    flow.FlowData             `json:"flow_source"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
	LastRunAt     null.Time                 `json:"last_run_at"`
}

type EventListenerListResponse []*EventListener

type LogEntryListResponse struct {
	Entries    []*wire.LogEntry `json:"entries"`
	NextBefore int64            `json:"next_before"`
}

func (h *SupportHandler) HandleAccessGet(c *handler.Context) (*AccessResponse, error) {
	return &AccessResponse{Authorized: true}, nil
}

func (h *SupportHandler) HandleAppGet(c *handler.Context) (*wire.AppGetResponse, error) {
	app, err := h.app(c)
	if err != nil {
		return nil, err
	}
	return wire.AppToWire(app), nil
}

func (h *SupportHandler) HandleLogEntryList(c *handler.Context) (*LogEntryListResponse, error) {
	beforeID := int64(0)
	if raw := c.Query("before"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 {
			return nil, handler.ErrBadRequest("invalid_before", "before must be a non-negative log ID")
		}
		beforeID = parsed
	}

	limit := 100
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			return nil, handler.ErrBadRequest("invalid_limit", "limit must be between 1 and 100")
		}
		limit = parsed
	}

	app, err := h.app(c)
	if err != nil {
		return nil, err
	}
	entries, err := h.logStore.LogEntriesByApp(c.Context(), app.ID, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get support log entries: %w", err)
	}

	levelFilter := c.Query("level")
	messageFilter := strings.ToLower(c.Query("query"))
	res := &LogEntryListResponse{Entries: make([]*wire.LogEntry, 0, len(entries))}
	for _, entry := range entries {
		if levelFilter != "" && !strings.EqualFold(string(entry.Level), levelFilter) {
			continue
		}
		if messageFilter != "" && !strings.Contains(strings.ToLower(entry.Message), messageFilter) {
			continue
		}
		res.Entries = append(res.Entries, wire.LogEntryToWire(entry))
	}

	if len(entries) == limit {
		res.NextBefore = entries[len(entries)-1].ID
	}
	return res, nil
}

func (h *SupportHandler) HandleCommandList(c *handler.Context) (*wire.CommandListResponse, error) {
	app, err := h.app(c)
	if err != nil {
		return nil, err
	}
	commands, err := h.commandStore.CommandsByApp(c.Context(), app.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get support commands: %w", err)
	}
	res := make([]*wire.Command, len(commands))
	for i, command := range commands {
		res[i] = wire.CommandToWire(command)
	}
	return &res, nil
}

func (h *SupportHandler) HandleEventListenerList(c *handler.Context) (*EventListenerListResponse, error) {
	app, err := h.app(c)
	if err != nil {
		return nil, err
	}
	listeners, err := h.eventListenerStore.EventListenersByApp(c.Context(), app.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get support event listeners: %w", err)
	}
	res := make(EventListenerListResponse, len(listeners))
	for i, listener := range listeners {
		res[i] = &EventListener{
			ID:            listener.ID,
			Source:        string(listener.Source),
			Type:          string(listener.Type),
			Description:   listener.Description,
			Enabled:       listener.Enabled,
			AppID:         listener.AppID,
			ModuleID:      listener.ModuleID,
			CreatorUserID: listener.CreatorUserID,
			Filter:        (*wire.EventListenerFilter)(listener.Filter),
			FlowSource:    listener.FlowSource,
			CreatedAt:     listener.CreatedAt,
			UpdatedAt:     listener.UpdatedAt,
			LastRunAt:     listener.LastRunAt,
		}
	}
	return &res, nil
}

func (h *SupportHandler) app(c *handler.Context) (*model.App, error) {
	appID := c.Param("appID")
	if appID == "" {
		return nil, handler.ErrBadRequest("missing_app_id", "Kite App ID is required")
	}
	app, err := h.appStore.App(c.Context(), appID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_app", "App not found")
		}
		return nil, fmt.Errorf("failed to get support app: %w", err)
	}
	return app, nil
}
