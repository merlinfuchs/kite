package editor_usage

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

var editorSessionIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type EditorUsageHandler struct {
	commandStore       store.CommandStore
	eventListenerStore store.EventListenerStore
	editorSessionStore store.EditorSessionStore
}

func NewEditorUsageHandler(
	commandStore store.CommandStore,
	eventListenerStore store.EventListenerStore,
	editorSessionStore store.EditorSessionStore,
) *EditorUsageHandler {
	return &EditorUsageHandler{
		commandStore:       commandStore,
		eventListenerStore: eventListenerStore,
		editorSessionStore: editorSessionStore,
	}
}

type EditorSessionUpsertRequest struct {
	TargetType    string `json:"target_type"`
	TargetID      string `json:"target_id"`
	ActiveSeconds int    `json:"active_seconds"`
	SaveCount     int    `json:"save_count"`
}

type EditorSessionUpsertResponse struct {
	Saved bool `json:"saved"`
}

type EditorUsageTotals struct {
	ActiveSeconds int64      `json:"active_seconds"`
	SaveCount     int64      `json:"save_count"`
	SessionCount  int64      `json:"session_count"`
	LastEditedAt  *time.Time `json:"last_edited_at"`
}

type EditorUsageResponse struct {
	Total EditorUsageTotals    `json:"total"`
	Items []*store.EditorUsage `json:"items"`
}

func (h *EditorUsageHandler) HandleEditorSessionUpsert(
	c *handler.Context,
	req EditorSessionUpsertRequest,
) (*EditorSessionUpsertResponse, error) {
	sessionID := c.Param("sessionID")
	if err := validateEditorSessionRequest(sessionID, req); err != nil {
		return nil, err
	}
	if c.App == nil || c.Session == nil {
		return nil, fmt.Errorf("editor session request missing app or user context")
	}

	if err := h.validateTarget(c, req.TargetType, req.TargetID); err != nil {
		return nil, err
	}

	err := h.editorSessionStore.UpsertEditorSession(c.Context(), &store.EditorSession{
		ID:            sessionID,
		AppID:         c.App.ID,
		TargetType:    req.TargetType,
		TargetID:      req.TargetID,
		UserID:        c.Session.UserID,
		ActiveSeconds: req.ActiveSeconds,
		SaveCount:     req.SaveCount,
	})
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, handler.ErrConflict("editor_session_conflict", "editor session does not match its original target")
		}
		return nil, fmt.Errorf("failed to record editor session: %w", err)
	}
	return &EditorSessionUpsertResponse{Saved: true}, nil
}

func validateEditorSessionRequest(sessionID string, req EditorSessionUpsertRequest) error {
	if !editorSessionIDPattern.MatchString(sessionID) {
		return handler.ErrBadRequest("invalid_editor_session", "session ID must be a UUID")
	}
	if req.TargetID == "" || len(req.TargetID) > 256 {
		return handler.ErrBadRequest("invalid_editor_target", "target ID is required")
	}
	if req.ActiveSeconds < 0 || req.ActiveSeconds > int((30*24*time.Hour).Seconds()) {
		return handler.ErrBadRequest("invalid_active_seconds", "active seconds must be between 0 and 30 days")
	}
	if req.SaveCount < 0 || req.SaveCount > 1_000_000 {
		return handler.ErrBadRequest("invalid_save_count", "save count must be between 0 and 1000000")
	}
	return nil
}

func (h *EditorUsageHandler) HandleEditorUsageGet(c *handler.Context) (*EditorUsageResponse, error) {
	if c.App == nil {
		return nil, fmt.Errorf("editor usage request missing app context")
	}
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	items, err := h.editorSessionStore.EditorUsageByApp(c.Context(), c.App.ID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get editor usage: %w", err)
	}

	res := &EditorUsageResponse{Items: items}
	for _, item := range items {
		res.Total.ActiveSeconds += item.ActiveSeconds
		res.Total.SaveCount += item.SaveCount
		res.Total.SessionCount += item.SessionCount
		if item.LastEditedAt != nil &&
			(res.Total.LastEditedAt == nil || item.LastEditedAt.After(*res.Total.LastEditedAt)) {
			res.Total.LastEditedAt = item.LastEditedAt
		}
	}
	return res, nil
}

func (h *EditorUsageHandler) validateTarget(c *handler.Context, targetType, targetID string) error {
	switch targetType {
	case "command":
		commands, err := h.commandStore.CommandsByApp(c.Context(), c.App.ID)
		if err != nil {
			return fmt.Errorf("failed to validate editor command: %w", err)
		}
		for _, command := range commands {
			if command.ID == targetID {
				return nil
			}
		}
	case "event_listener":
		listeners, err := h.eventListenerStore.EventListenersByApp(c.Context(), c.App.ID)
		if err != nil {
			return fmt.Errorf("failed to validate editor event listener: %w", err)
		}
		for _, listener := range listeners {
			if listener.ID == targetID {
				return nil
			}
		}
	default:
		return handler.ErrBadRequest("invalid_editor_target", "target type must be command or event_listener")
	}
	return handler.ErrBadRequest("invalid_editor_target", "target does not belong to this app")
}
