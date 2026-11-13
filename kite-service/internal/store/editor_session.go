package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

const (
	// EditorSessionTimeSlack is how much more active time than the age of a
	// session the editor can report, for the delay of its requests.
	EditorSessionTimeSlack = 30 * time.Second
	// EditorSessionMaxTime caps the active time of one session.
	EditorSessionMaxTime = 12 * time.Hour
)

type EditorSessionStore interface {
	// UpdateEditorSessionTime sets the active time of the session, starting
	// it if it's new. See EditorSessionTimeSlack and EditorSessionMaxTime.
	UpdateEditorSessionTime(ctx context.Context, session model.EditorSession, activeSeconds int, now time.Time) error
	// RecordEditorSessionSave counts a save of the session, starting it if
	// it's new.
	RecordEditorSessionSave(ctx context.Context, session model.EditorSession, now time.Time) error
	// EditorUsageByApp sums the sessions started since since per command and
	// event listener.
	EditorUsageByApp(ctx context.Context, appID string, since time.Time) ([]*model.EditorUsage, error)
	// DeleteOldEditorSessions deletes up to batchSize sessions started before
	// createdBefore.
	DeleteOldEditorSessions(ctx context.Context, createdBefore time.Time, batchSize int) (int64, error)
}
