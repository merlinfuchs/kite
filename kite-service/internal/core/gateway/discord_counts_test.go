package gateway

import (
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func TestCountsDueAt(t *testing.T) {
	if due := countsDueAt(&model.App{}); !due.IsZero() {
		t.Errorf("never counted: due at %v, want right away", due)
	}

	updated := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	app := &model.App{DiscordCountsUpdatedAt: null.TimeFrom(updated)}
	if due, want := countsDueAt(app), updated.Add(countsInterval); !due.Equal(want) {
		t.Errorf("due at %v, want %v", due, want)
	}
}

func TestDueForCounts(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	gateway := func(id string, dueAt time.Time) *Gateway {
		return &Gateway{appID: id, countsDueAt: dueAt}
	}

	gateways := []*Gateway{
		gateway("later", now.Add(time.Minute)),
		gateway("now", now),
		gateway("never", time.Time{}),
		gateway("hour", now.Add(-time.Hour)),
	}

	var ids []string
	for _, g := range dueForCounts(gateways, now, 2) {
		ids = append(ids, g.appID)
	}
	if len(ids) != 2 || ids[0] != "never" || ids[1] != "hour" {
		t.Errorf("due = %v, want [never hour]", ids)
	}

	if due := dueForCounts(gateways, now, 10); len(due) != 3 {
		t.Errorf("got %d due, want 3", len(due))
	}
}
