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
