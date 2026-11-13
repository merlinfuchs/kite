package editorusage

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// usagePeriod is how far back the editor usage of an app goes.
const usagePeriod = 30 * 24 * time.Hour

type EditorUsageHandler struct {
	editorSessionStore store.EditorSessionStore
}

func NewEditorUsageHandler(editorSessionStore store.EditorSessionStore) *EditorUsageHandler {
	return &EditorUsageHandler{
		editorSessionStore: editorSessionStore,
	}
}

func (h *EditorUsageHandler) HandleCommandEditorSessionUpdate(c *handler.Context, req wire.EditorSessionUpdateRequest) (*wire.EditorSessionUpdateResponse, error) {
	return h.updateSession(c, CommandSession(c, c.Param("sessionID")), req)
}

func (h *EditorUsageHandler) HandleEventListenerEditorSessionUpdate(c *handler.Context, req wire.EditorSessionUpdateRequest) (*wire.EditorSessionUpdateResponse, error) {
	return h.updateSession(c, EventListenerSession(c, c.Param("sessionID")), req)
}

func (h *EditorUsageHandler) updateSession(c *handler.Context, session model.EditorSession, req wire.EditorSessionUpdateRequest) (*wire.EditorSessionUpdateResponse, error) {
	if session.ID == "" || len(session.ID) > 64 {
		return nil, handler.ErrBadRequest("invalid_editor_session", "Editor session ID must be 1 to 64 characters")
	}

	err := h.editorSessionStore.UpdateEditorSessionTime(c.Context(), session, req.ActiveSeconds, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to update editor session: %w", err)
	}
	return &wire.EditorSessionUpdateResponse{}, nil
}

func (h *EditorUsageHandler) HandleEditorUsageGet(c *handler.Context) (*wire.EditorUsageGetResponse, error) {
	usage, err := h.editorSessionStore.EditorUsageByApp(c.Context(), c.App.ID, time.Now().UTC().Add(-usagePeriod))
	if err != nil {
		return nil, fmt.Errorf("failed to get editor usage: %w", err)
	}

	res := &wire.EditorUsageGetResponse{
		Commands:       map[string]*wire.EditorUsage{},
		EventListeners: map[string]*wire.EditorUsage{},
	}
	for _, u := range usage {
		item := wire.EditorUsageToWire(u)
		res.Total.Add(item)
		if u.CommandID.Valid {
			res.Commands[u.CommandID.String] = item
		} else {
			res.EventListeners[u.EventListenerID.String] = item
		}
	}
	return res, nil
}

// CommandSession is the editor session of the request's command.
func CommandSession(c *handler.Context, sessionID string) model.EditorSession {
	return model.EditorSession{
		ID:        sessionID,
		AppID:     c.App.ID,
		UserID:    c.Session.UserID,
		CommandID: null.StringFrom(c.Command.ID),
	}
}

// EventListenerSession is the editor session of the request's event listener.
func EventListenerSession(c *handler.Context, sessionID string) model.EditorSession {
	return model.EditorSession{
		ID:              sessionID,
		AppID:           c.App.ID,
		UserID:          c.Session.UserID,
		EventListenerID: null.StringFrom(c.EventListener.ID),
	}
}

// RecordSave counts a save made in an editor. Saves made with an API token or
// without an editor session, like restores, aren't counted. It only logs
// errors, as the save itself succeeded.
func RecordSave(c *handler.Context, sessions store.EditorSessionStore, session model.EditorSession) {
	if c.APIToken != nil || session.ID == "" {
		return
	}
	if err := sessions.RecordEditorSessionSave(c.Context(), session, time.Now().UTC()); err != nil {
		slog.Error(
			"Failed to record editor save",
			slog.String("app_id", session.AppID),
			slog.String("error", err.Error()),
		)
	}
}
