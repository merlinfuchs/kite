package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/merlinfuchs/dismod/disrest"
	"github.com/merlinfuchs/dismod/disway"
	"github.com/merlinfuchs/kite/kite-service/internal/bot/state"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type Bot struct {
	Cluster    *disway.Cluster
	Client     *disrest.Client
	Engine     *engine.Engine
	State      *state.BotState
	guildStore store.GuildStore
}

func New(token string, pg *postgres.Client) (*Bot, error) {
	client := disrest.NewClient(token, slog.Default())

	gInfo, err := client.GatewayBot()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway: %w", err)
	}

	shardIDs := make([]int, gInfo.Shards)
	for i := range shardIDs {
		shardIDs[i] = i
	}

	cluster := disway.NewCluster(token, slog.Default(),
		sharding.WithShardCount(gInfo.Shards),
		sharding.WithShardIDs(shardIDs...),
		sharding.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds|gateway.IntentGuildMessages), // | gateway.IntentGuildMembers),
			gateway.WithPresenceOpts(
				gateway.WithCustomActivity("kite.only"),
			),
		),
	)

	state := state.New(client)
	cluster.AddAllEventListener(state.Update)

	b := &Bot{
		Client:     client,
		Cluster:    cluster,
		State:      state,
		guildStore: pg,
	}

	b.registerListeners()

	return b, nil
}

func (b *Bot) Open(ctx context.Context) {
	b.Cluster.Open(ctx)
}

func (b *Bot) Close(ctx context.Context) {
	b.Cluster.Close(ctx)
}
