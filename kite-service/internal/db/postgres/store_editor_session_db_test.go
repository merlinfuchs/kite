package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestEditorSessions(t *testing.T) {
	c := testDB(t)
	ctx := context.Background()
	_, err := c.DB.Exec(ctx, `
INSERT INTO users (id, email, display_name, discord_id, discord_username, created_at, updated_at)
VALUES ('test_user', 'test@example.com', 'Test', 'test_user', 'test', now(), now()) ON CONFLICT DO NOTHING;
INSERT INTO apps (id, name, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at)
VALUES ('test_app', 'Test', 'test_user', 'test_user', 't', 'test_app', now(), now()) ON CONFLICT DO NOTHING;
INSERT INTO commands (id, name, description, app_id, creator_user_id, flow_source, enabled, created_at, updated_at)
VALUES ('test_cmd', 'test', 'Test', 'test_app', 'test_user', '{}', true, now(), now()) ON CONFLICT DO NOTHING;
DELETE FROM editor_sessions WHERE app_id = 'test_app';`)
	require.NoError(t, err)

	session := model.EditorSession{
		ID:        "test_session",
		AppID:     "test_app",
		UserID:    "test_user",
		CommandID: null.StringFrom("test_cmd"),
	}
	start := time.Now().UTC().Truncate(time.Second)
	usage := func() *model.EditorUsage {
		res, err := c.EditorUsageByApp(ctx, "test_app", start.Add(-time.Hour))
		require.NoError(t, err)
		require.Len(t, res, 1)
		return res[0]
	}

	// A new session can't start with more time than the slack.
	require.NoError(t, c.UpdateEditorSessionTime(ctx, session, 3600, start))
	require.Equal(t, 30, usage().ActiveSeconds)

	// Later it can't have more time than its age plus the slack.
	require.NoError(t, c.UpdateEditorSessionTime(ctx, session, 3600, start.Add(time.Minute)))
	require.Equal(t, 90, usage().ActiveSeconds)

	// Less time than before is ignored.
	require.NoError(t, c.UpdateEditorSessionTime(ctx, session, 10, start.Add(2*time.Minute)))
	require.Equal(t, 90, usage().ActiveSeconds)

	require.NoError(t, c.UpdateEditorSessionTime(ctx, session, 100, start.Add(2*time.Minute)))
	require.Equal(t, 100, usage().ActiveSeconds)

	// No session has more than the max time.
	require.NoError(t, c.UpdateEditorSessionTime(ctx, session, 1_000_000_000, start.Add(30*24*time.Hour)))
	require.Equal(t, 12*60*60, usage().ActiveSeconds)

	require.NoError(t, c.RecordEditorSessionSave(ctx, session, start.Add(3*time.Minute)))
	require.NoError(t, c.RecordEditorSessionSave(ctx, session, start.Add(4*time.Minute)))
	u := usage()
	require.Equal(t, 2, u.SaveCount)
	require.Equal(t, start.Add(4*time.Minute), u.LastSavedAt.Time)

	// The ID of a session of another user or item changes nothing.
	other := session
	other.CommandID = null.String{}
	other.EventListenerID = null.StringFrom("other")
	require.NoError(t, c.RecordEditorSessionSave(ctx, other, start.Add(5*time.Minute)))
	require.Equal(t, 2, usage().SaveCount)

	deleted, err := c.DeleteOldEditorSessions(ctx, start.Add(time.Second), 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted)
}
