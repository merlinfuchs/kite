package postgres

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestMergesInto(t *testing.T) {
	now := time.Now().UTC()
	newest := pgmodel.GetNewestItemVersionRow{
		Source:        string(model.ItemVersionSourceEditor),
		CreatorUserID: pgtype.Text{String: "user", Valid: true},
		MergeKey:      pgtype.Text{String: "session", Valid: true},
		CreatedAt:     pgtype.Timestamp{Time: now.Add(-time.Minute), Valid: true},
	}
	expired := newest
	expired.CreatedAt = pgtype.Timestamp{Time: now.Add(-store.EditorSessionMergeWindow - time.Second), Valid: true}
	unkeyed := newest
	unkeyed.MergeKey = pgtype.Text{}

	tests := []struct {
		name   string
		newest pgmodel.GetNewestItemVersionRow
		change store.ItemChange
		want   bool
	}{
		{"same session", newest, store.EditorChange("user", "session"), true},
		{"other session", newest, store.EditorChange("user", "other"), false},
		{"restore without session", newest, store.EditorChange("user", ""), false},
		{"after a version without key", unkeyed, store.EditorChange("user", "session"), false},
		{"other user", newest, store.EditorChange("other", "session"), false},
		{"other source with same key", newest, store.APITokenChange("user", "session"), false},
		{"window over", expired, store.EditorChange("user", "session"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, mergesInto(tt.newest, tt.change, now))
		})
	}
}

func TestContentVersion(t *testing.T) {
	row := pgmodel.Command{FlowSource: []byte(`{"nodes": []}`)}
	version := commandVersion(row)

	toggled := row
	toggled.Enabled = true
	toggled.UpdatedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
	assert.Equal(t, version, commandVersion(toggled), "not changed by enabling")

	changed := row
	changed.FlowSource = []byte(`{"nodes": [], "edges": []}`)
	assert.NotEqual(t, version, commandVersion(changed))

	// Parts can't run into each other.
	assert.NotEqual(t, contentVersion([]byte("ab"), []byte("c")), contentVersion([]byte("a"), []byte("bc")))
}

func TestCheckBaseVersion(t *testing.T) {
	change := store.EditorChange("user", "session")
	assert.NoError(t, checkBaseVersion("v1", change), "no base")

	change.BaseVersion = "v1"
	assert.NoError(t, checkBaseVersion("v1", change))
	assert.ErrorIs(t, checkBaseVersion("v2", change), store.ErrOutdated)
}
