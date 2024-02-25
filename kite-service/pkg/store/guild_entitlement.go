package store

import (
	"context"
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type GuildEntitlementStore interface {
	// UpsertGuildEntitlement creates a new entitlement or updates an existing one if (guild_id, source, source_id) match
	UpsertGuildEntitlement(ctx context.Context, entilement model.GuildEntitlement) (*model.GuildEntitlement, error)
	GetGuildEntitlements(ctx context.Context, guildID distype.Snowflake, validAt time.Time) ([]model.GuildEntitlement, error)
	GetResolvedGuildEntitlement(ctx context.Context, guildID distype.Snowflake) (*model.GuildEntitlementResolved, error)
}
