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

type EditorUsageHandler struct {
	editorSessionStore store.EditorSessionStore
}

func NewEditorUsageHandler(editorSessionStore store.EditorSessionStore) *EditorUsageHandler {
	return &EditorUsageHandler{
		editorSessionStore: editorSessionStore,
	}
}

// HandleEditorSessionUpdate is routed under the command and event listener
// routes.
func (h *EditorUsageHandler) HandleEditorSessionUpdate(c *handler.Context, req wire.EditorSessionUpdateRequest) (*wire.EditorSessionUpdateResponse, error) {
	session := ItemSession(c, c.Param("sessionID"))
	if len(session.ID) > wire.EditorSessionMaxLength {
		return nil, handler.ErrBadRequest("invalid_editor_session", "Editor session ID is too long")
	}

	err := h.editorSessionStore.UpdateEditorSessionTime(c.Context(), session, req.ActiveSeconds, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to update editor session: %w", err)
	}
	return &wire.EditorSessionUpdateResponse{}, nil
}

func (h *EditorUsageHandler) HandleEditorUsageGet(c *handler.Context) (*wire.EditorUsageGetResponse, error) {
	usage, err := h.editorSessionStore.EditorUsageByApp(c.Context(), c.App.ID, time.Now().UTC().Add(-store.EditorUsagePeriod))
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

// ItemSession is the editor session of the request's command or event
// listener.
func ItemSession(c *handler.Context, sessionID string) model.EditorSession {
	session := model.EditorSession{
		ID:     sessionID,
		AppID:  c.App.ID,
		UserID: c.Session.UserID,
	}
	if c.Command != nil {
		session.CommandID = null.StringFrom(c.Command.ID)
	} else {
		session.EventListenerID = null.StringFrom(c.EventListener.ID)
	}
	return session
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
