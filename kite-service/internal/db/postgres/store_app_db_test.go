package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Whoever switches a shared app becomes its owner, without the collaborators
// and invites it had. The old owner stays as an admin while they pay for it.
func TestSwitchSharedAppToOwn(t *testing.T) {
	c := testDB(t)
	ctx := context.Background()
	_, err := c.DB.Exec(ctx, `
INSERT INTO users (id, email, display_name, discord_id, discord_username, created_at, updated_at) VALUES
('switch_owner', 'owner@example.com', 'Owner', 'switch_owner', 'owner', now(), now()),
('switch_collab', 'collab@example.com', 'Collab', 'switch_collab', 'collab', now(), now()),
('switch_user', 'user@example.com', 'User', 'switch_user', 'user', now(), now())
ON CONFLICT DO NOTHING;
DELETE FROM apps WHERE id = 'switch_app';
INSERT INTO apps (id, name, owner_user_id, creator_user_id, bot_type, discord_guild_id, created_at, updated_at)
VALUES ('switch_app', 'Server', 'switch_owner', 'switch_owner', 'shared', 'switch_guild', now(), now());
INSERT INTO collaborators (user_id, app_id, role, created_at, updated_at)
VALUES ('switch_collab', 'switch_app', 'admin', now(), now());
INSERT INTO collaborator_invites (app_id, discord_username, role, created_at, expires_at)
VALUES ('switch_app', 'someone', 'admin', now(), now() + interval '1 day');
DELETE FROM subscriptions WHERE id = 'switch_sub';
INSERT INTO subscriptions (id, display_name, source, status, status_formatted, created_at, updated_at, user_id)
VALUES ('switch_sub', 'Plus', 'lemonsqueezy', 'active', 'Active', now(), now(), 'switch_owner');
INSERT INTO entitlements (id, type, subscription_id, app_id, plan_id, created_at, updated_at)
VALUES ('switch_ent', 'subscription', 'switch_sub', 'switch_app', 'plus', now(), now());`)
	require.NoError(t, err)

	app, err := c.SwitchSharedAppToOwn(ctx, store.SharedAppSwitchOpts{
		ID:           "switch_app",
		OwnerUserID:  "switch_user",
		Name:         "Bot",
		DiscordID:    "switch_bot",
		DiscordToken: "t",
		UpdatedAt:    time.Now(),
	})
	require.NoError(t, err)
	assert.Equal(t, "switch_user", app.OwnerUserID)
	assert.False(t, app.IsShared())

	collaborators := func() []string {
		rows, err := c.DB.Query(ctx, `SELECT user_id || ':' || role FROM collaborators WHERE app_id = 'switch_app' ORDER BY user_id`)
		require.NoError(t, err)
		var res []string
		for rows.Next() {
			var v string
			require.NoError(t, rows.Scan(&v))
			res = append(res, v)
		}
		return res
	}
	var invites int
	require.NoError(t, c.DB.QueryRow(ctx, `SELECT count(*) FROM collaborator_invites WHERE app_id = 'switch_app'`).Scan(&invites))
	assert.Equal(t, []string{"switch_owner:admin"}, collaborators())
	assert.Zero(t, invites)

	// Switching it again changes nothing, also not the collaborators the new
	// owner added.
	_, err = c.DB.Exec(ctx, `INSERT INTO collaborators (user_id, app_id, role, created_at, updated_at)
VALUES ('switch_collab', 'switch_app', 'admin', now(), now())`)
	require.NoError(t, err)
	_, err = c.SwitchSharedAppToOwn(ctx, store.SharedAppSwitchOpts{
		ID:           "switch_app",
		OwnerUserID:  "switch_collab",
		Name:         "Bot",
		DiscordID:    "switch_bot_2",
		DiscordToken: "t",
		UpdatedAt:    time.Now(),
	})
	assert.ErrorIs(t, err, store.ErrNotFound)
	var owner string
	require.NoError(t, c.DB.QueryRow(ctx, `SELECT owner_user_id FROM apps WHERE id = 'switch_app'`).Scan(&owner))
	assert.Equal(t, "switch_user", owner)
	assert.Equal(t, []string{"switch_collab:admin", "switch_owner:admin"}, collaborators())
}
