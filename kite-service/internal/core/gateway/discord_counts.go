package gateway

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// The counts of servers and users that installed an app come from the REST
// API, as Discord has no gateway event for user installs. Each app is
// refreshed about once a day, at most countsBatchSize a minute on each
// cluster, so apps that were never counted don't all send a request at once.
const (
	countsInterval   = 24 * time.Hour
	countsRetryDelay = time.Hour
	countsBatchSize  = 100
	countsTimeout    = 10 * time.Second
)

// countsDueAt returns when the counts of the app are refreshed next, which is
// right away if they never were.
func countsDueAt(app *model.App) time.Time {
	if !app.DiscordCountsUpdatedAt.Valid {
		return time.Time{}
	}
	return app.DiscordCountsUpdatedAt.Time.Add(countsInterval)
}

// refreshCounts stores Discord's approximate counts of the servers and users
// that installed the app.
func (g *Gateway) refreshCounts(ctx context.Context, now time.Time) {
	g.countsDueAt = now.Add(countsRetryDelay)

	ctx, cancel := context.WithTimeout(ctx, countsTimeout)
	defer cancel()

	app, err := g.Session().Client.WithContext(ctx).CurrentApplication()
	if err != nil {
		slog.Warn(
			"Failed to get app counts from Discord",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
		return
	}

	err = g.appStore.UpdateAppDiscordCounts(ctx, store.AppDiscordCountsUpdateOpts{
		ID:               g.appID,
		GuildCount:       app.ApproximateGuildCount,
		UserInstallCount: app.ApproximateUserInstallCount,
		UpdatedAt:        now.UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to update app counts",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
		return
	}

	g.countsDueAt = now.Add(countsInterval)
}

// dueForCounts returns up to limit gateways whose counts are due, the longest
// overdue first.
func dueForCounts(gateways []*Gateway, now time.Time, limit int) []*Gateway {
	var due []*Gateway
	for _, g := range gateways {
		if !g.countsDueAt.After(now) {
			due = append(due, g)
		}
	}

	slices.SortFunc(due, func(a, b *Gateway) int {
		return a.countsDueAt.Compare(b.countsDueAt)
	})
	if len(due) > limit {
		due = due[:limit]
	}
	return due
}

func (m *GatewayManager) refreshCounts(ctx context.Context, now time.Time) {
	m.Lock()
	gateways := slices.Collect(maps.Values(m.gateways))
	m.Unlock()

	for _, g := range dueForCounts(gateways, now, countsBatchSize) {
		if ctx.Err() != nil {
			return
		}
		g.refreshCounts(ctx, now)
	}
}
