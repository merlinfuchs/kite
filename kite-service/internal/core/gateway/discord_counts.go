package gateway

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"time"
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

func (m *GatewayManager) refreshCounts(ctx context.Context, now time.Time) {
	m.Lock()
	gateways := slices.Collect(maps.Values(m.gateways))
	m.Unlock()

	var refreshed int
	for _, g := range gateways {
		if refreshed == countsBatchSize || ctx.Err() != nil {
			return
		}
		if g.countsDueAt.After(now) {
			continue
		}
		g.refreshCounts(ctx, now)
		refreshed++
	}
}

// refreshCounts stores Discord's approximate counts of the servers and users
// that installed the app.
func (g *Gateway) refreshCounts(ctx context.Context, now time.Time) {
	g.countsDueAt = now.Add(countsRetryDelay)

	discordCtx, cancel := context.WithTimeout(ctx, countsTimeout)
	defer cancel()

	app, err := g.Session().Client.WithContext(discordCtx).CurrentApplication()
	if err != nil {
		slog.Warn(
			"Failed to get app counts from Discord",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
		return
	}

	err = g.appStore.UpdateAppDiscordCounts(ctx, g.appID, app.ApproximateGuildCount, app.ApproximateUserInstallCount)
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
