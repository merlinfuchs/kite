package store

import (
	"context"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type GuildUsageStore interface {
	CreateGuildUsageEntry(ctx context.Context, entry model.GuildUsageEntry) error
	GetLastGuildUsageEntry(ctx context.Context, guildID distype.Snowflake) (*model.GuildUsageEntry, error)
	GetGuildUsageSummary(ctx context.Context, guildID distype.Snowflake) (*model.GuildUsageSummary, error)
	GetGuildUsageAndLimits(ctx context.Context, guildID distype.Snowflake) (*model.GuildUsageAndLimits, error)
}
