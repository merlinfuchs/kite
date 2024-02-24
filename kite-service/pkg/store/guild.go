package store

import (
	"context"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type GuildStore interface {
	UpsertGuild(ctx context.Context, deployment model.Guild) (*model.Guild, error)
	GetGuilds(ctx context.Context) ([]model.Guild, error)
	GetGuild(ctx context.Context, id distype.Snowflake) (*model.Guild, error)
	GetDistinctGuildIDs(ctx context.Context) ([]distype.Snowflake, error)
}
