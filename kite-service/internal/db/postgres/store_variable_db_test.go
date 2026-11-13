package postgres

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// testDB connects to the database in KITE_TEST_POSTGRES, a DSN of an empty
// database it migrates, like a postgres container. Tests that need one skip
// without it.
func testDB(t *testing.T) *Client {
	dsn := os.Getenv("KITE_TEST_POSTGRES")
	if dsn == "" {
		t.Skip("KITE_TEST_POSTGRES isn't set")
	}
	c, err := New(dsn, 1)
	require.NoError(t, err)
	mig, err := c.GetMigrater()
	require.NoError(t, err)
	_ = mig.Up()
	t.Cleanup(c.DB.Close)
	return c
}

// Updates of a value that doesn't exist yet each count, also when runs make
// them at the same time.
func TestUpdateVariableValueConcurrently(t *testing.T) {
	c := testDB(t)
	ctx := context.Background()
	_, err := c.DB.Exec(ctx, `
INSERT INTO users (id, email, display_name, discord_id, discord_username, created_at, updated_at)
VALUES ('test_user', 'test@example.com', 'Test', 'test_user', 'test', now(), now()) ON CONFLICT DO NOTHING;
INSERT INTO apps (id, name, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at)
VALUES ('test_app', 'Test', 'test_user', 'test_user', 't', 'test_app', now(), now()) ON CONFLICT DO NOTHING;
INSERT INTO variables (id, name, scoped, app_id, created_at, updated_at)
VALUES ('test_var', 'test', true, 'test_app', now(), now()) ON CONFLICT DO NOTHING;
DELETE FROM variable_values WHERE variable_id = 'test_var';`)
	require.NoError(t, err)

	update := func(op provider.VariableOperation, scope string, v thing.Thing) error {
		_, err := c.UpdateVariableValue(ctx, "test_app", op, model.VariableValue{
			VariableID: "test_var", Scope: null.StringFrom(scope), Data: v, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		})
		return err
	}
	value := func(scope string) thing.Thing {
		v, err := c.VariableValue(ctx, "test_app", "test_var", null.StringFrom(scope))
		require.NoError(t, err)
		return v.Data
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.NoError(t, update(provider.VariableOperationIncrement, "inc", thing.NewInt(1)))
		}()
		go func() {
			defer wg.Done()
			assert.NoError(t, update(provider.VariableOperationAppend, "list", thing.NewObject(nil)))
		}()
	}
	wg.Wait()
	assert.Equal(t, float64(20), value("inc").Float())
	assert.Len(t, value("list").Array(), 20)

	// A variable of another app isn't changed.
	_, err = c.UpdateVariableValue(ctx, "other_app", provider.VariableOperationIncrement, model.VariableValue{
		VariableID: "test_var", Scope: null.StringFrom("inc"), Data: thing.NewInt(1), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	assert.ErrorIs(t, err, store.ErrNotFound)
}
