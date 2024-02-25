package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

func (c *Client) UpsertGuildEntitlement(ctx context.Context, entitlement model.GuildEntitlement) (*model.GuildEntitlement, error) {
	row, err := c.Q.UpsertGuildEntitlement(ctx, pgmodel.UpsertGuildEntitlementParams{
		ID:      entitlement.ID,
		GuildID: string(entitlement.GuildID),
		UserID: pgtype.Text{
			String: string(entitlement.UserID.Value),
			Valid:  entitlement.UserID.Valid,
		},
		Source:                        string(entitlement.Source),
		SourceID:                      nullStringToText(entitlement.SourceID),
		Name:                          nullStringToText(entitlement.Name),
		Description:                   nullStringToText(entitlement.Description),
		FeatureMonthlyCpuTimeLimit:    int32(entitlement.Features.MonthlyCpuTimeLimit.Milliseconds()),
		FeatureMonthlyCpuTimeAdditive: entitlement.Features.MonthlyCpuTimeAdditive,
		CreatedAt:                     timeToTimestamp(entitlement.CreatedAt),
		UpdatedAt:                     timeToTimestamp(entitlement.UpdatedAt),
		ValidFrom:                     nullTimeToTimestamp(entitlement.ValidFrom),
		ValidUntil:                    nullTimeToTimestamp(entitlement.ValidUntil),
	})
	if err != nil {
		return nil, err
	}

	res := guildEntitlementToModel(row)
	return &res, nil
}

func (c *Client) GetGuildEntitlements(ctx context.Context, guildID distype.Snowflake, validAt time.Time) ([]model.GuildEntitlement, error) {
	rows, err := c.Q.GetGuildEntitlements(ctx, pgmodel.GetGuildEntitlementsParams{
		GuildID: string(guildID),
		ValidAt: timeToTimestamp(validAt),
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.GuildEntitlement, len(rows))
	for i, row := range rows {
		res[i] = guildEntitlementToModel(row)
	}

	return res, nil
}

func (c *Client) GetResolvedGuildEntitlement(ctx context.Context, guildID distype.Snowflake) (*model.GuildEntitlementResolved, error) {
	row, err := c.Q.GetResolvedGuildEntitlement(ctx, string(guildID))
	if err != nil {
		return nil, err
	}

	return &model.GuildEntitlementResolved{
		MonthlyCpuTimeLimit: int(row.FeatureMonthlyCpuTimeLimit),
	}, nil
}

func guildEntitlementToModel(row pgmodel.GuildEntitlement) model.GuildEntitlement {
	return model.GuildEntitlement{
		ID:      row.ID,
		GuildID: distype.Snowflake(row.GuildID),
		UserID: distype.Nullable[distype.Snowflake]{
			Value: distype.Snowflake(row.UserID.String),
			Valid: row.UserID.Valid,
		},
		Source:      model.GuildEntitlementSource(row.Source),
		SourceID:    textToNullString(row.SourceID),
		Name:        textToNullString(row.Name),
		Description: textToNullString(row.Description),
		Features: model.GuildEntitlementFeatures{
			MonthlyCpuTimeLimit:    time.Duration(row.FeatureMonthlyCpuTimeLimit) * time.Millisecond,
			MonthlyCpuTimeAdditive: row.FeatureMonthlyCpuTimeAdditive,
		},
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
		ValidFrom:  timestampToNullTime(row.ValidFrom),
		ValidUntil: timestampToNullTime(row.ValidUntil),
	}
}
