package postgres

import (
	"encoding/json"
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

func TestCheckChangeBase(t *testing.T) {
	// Postgres keeps microseconds, and clients send back the updated_at they
	// got as JSON.
	updatedAt := time.Date(2026, 10, 27, 12, 0, 0, 123456000, time.UTC)
	raw, err := json.Marshal(updatedAt)
	assert.NoError(t, err)
	var sent time.Time
	assert.NoError(t, json.Unmarshal(raw, &sent))

	change := store.EditorChange("user", "session")
	assert.NoError(t, checkChangeBase(updatedAt, change), "no base")

	change.BaseUpdatedAt = sent
	assert.NoError(t, checkChangeBase(updatedAt, change), "same version")

	change.BaseUpdatedAt = sent.Add(-time.Second)
	assert.ErrorIs(t, checkChangeBase(updatedAt, change), store.ErrOutdated)
}
