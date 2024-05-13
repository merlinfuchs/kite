package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/merlinfuchs/dismod/disrest"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/dismod/disway"
	"github.com/merlinfuchs/kite/kite-service/internal/app/state"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.AppStateProvider = (*App)(nil)

type App struct {
	appID distype.Snowflake

	cluster *disway.Cluster
	client  *disrest.Client
	state   *state.AppState

	appUsageStore store.AppUsageStore
	engine        *engine.Engine
}

func NewApp(
	appID distype.Snowflake,
	info *model.App,
	appUsageStore store.AppUsageStore,
	engine *engine.Engine,
) (*App, error) {
	g := &App{
		appID:         appID,
		appUsageStore: appUsageStore,
		engine:        engine,
	}

	// TODO: use proxy
	client := disrest.NewClient(info.Token, slog.Default())

	gInfo, err := client.GatewayBot()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway: %w", err)
	}

	if gInfo.Shards > gInfo.SessionStartLimit.Remaining {
		return nil, fmt.Errorf("remaining session start limit is too low")
	}

	shardIDs := make([]int, gInfo.Shards)
	for i := range shardIDs {
		shardIDs[i] = i
	}

	cluster := disway.NewCluster(info.Token, slog.Default(),
		sharding.WithShardCount(gInfo.Shards),
		sharding.WithShardIDs(shardIDs...),
		sharding.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds|gateway.IntentGuildMessages|gateway.IntentMessageContent), // | gateway.IntentGuildMembers),
			gateway.WithPresenceOpts(
				gateway.WithCustomActivity(info.StatusActivityName.String),
			),
		),
	)

	state := state.New(client)
	cluster.AddAllEventListener(state.Update)

	g.cluster = cluster
	g.client = client
	g.state = state

	g.registerListeners()

	return g, nil
}

func (a *App) Open(ctx context.Context) {
	a.cluster.Open(ctx)
}

func (a *App) Close(ctx context.Context) {
	a.cluster.Close(ctx)
	a.client.Close(ctx)
}

func (a *App) State() store.DiscordStateStore {
	return a.state
}

func (a *App) Client() *disrest.Client {
	return a.client
}
