package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A command restored with its old ID gets a version after its delete, so its
// history isn't deleted with the ones of deleted items.
func TestRestoreBundleContentsKeepsHistory(t *testing.T) {
	c := testDB(t)
	ctx := context.Background()
	_, err := c.DB.Exec(ctx, `
INSERT INTO users (id, email, display_name, discord_id, discord_username, created_at, updated_at)
VALUES ('restore_user', 'restore@example.com', 'Restore', 'restore_user', 'restore', now(), now()) ON CONFLICT DO NOTHING;
INSERT INTO apps (id, name, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at)
VALUES ('restore_app', 'Restore', 'restore_user', 'restore_user', 't', 'restore_app', now(), now()) ON CONFLICT DO NOTHING;
DELETE FROM commands WHERE id = 'restore_cmd';
DELETE FROM item_versions WHERE target_id = 'restore_cmd';`)
	require.NoError(t, err)

	now := time.Now().UTC()
	cmd := &model.Command{
		ID:            "restore_cmd",
		Name:          "ping",
		AppID:         "restore_app",
		CreatorUserID: "restore_user",
		FlowSource:    flow.FlowData{Nodes: []flow.FlowNode{{ID: "entry", Type: flow.FlowNodeTypeEntryCommand}}},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err = c.CreateCommand(ctx, cmd)
	require.NoError(t, err)
	require.NoError(t, c.DeleteCommand(ctx, cmd.ID, "restore_user"))

	require.NoError(t, c.RestoreBundleContents(ctx, store.BundleInstallContents{Commands: []*model.Command{cmd}}, "restore_user"))

	newest, err := c.Q.GetNewestItemVersion(ctx, cmd.ID)
	require.NoError(t, err)
	assert.Equal(t, string(model.ItemVersionSourceBundle), newest.Source)

	_, err = c.DeleteDeletedItemVersions(ctx, time.Now().Add(time.Hour), 100)
	require.NoError(t, err)
	var count int
	require.NoError(t, c.DB.QueryRow(ctx, `SELECT count(*) FROM item_versions WHERE target_id = 'restore_cmd'`).Scan(&count))
	assert.Equal(t, 2, count)
}
