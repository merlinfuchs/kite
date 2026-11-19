package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Whoever switches a shared app becomes its only owner, without the owner,
// collaborators and invites it had.
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
VALUES ('switch_app', 'someone', 'admin', now(), now() + interval '1 day');`)
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

	var collaborators, invites int
	require.NoError(t, c.DB.QueryRow(ctx, `SELECT count(*) FROM collaborators WHERE app_id = 'switch_app'`).Scan(&collaborators))
	require.NoError(t, c.DB.QueryRow(ctx, `SELECT count(*) FROM collaborator_invites WHERE app_id = 'switch_app'`).Scan(&invites))
	assert.Zero(t, collaborators)
	assert.Zero(t, invites)

	// An app that isn't shared anymore can't be switched again.
	_, err = c.SwitchSharedAppToOwn(ctx, store.SharedAppSwitchOpts{
		ID: "switch_app", OwnerUserID: "switch_collab", Name: "Bot", DiscordID: "switch_bot_2", DiscordToken: "t", UpdatedAt: time.Now(),
	})
	assert.ErrorIs(t, err, store.ErrNotFound)
	var owner string
	require.NoError(t, c.DB.QueryRow(ctx, `SELECT owner_user_id FROM apps WHERE id = 'switch_app'`).Scan(&owner))
	assert.Equal(t, "switch_user", owner)
}
