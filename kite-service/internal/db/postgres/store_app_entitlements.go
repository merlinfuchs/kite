package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.AppEntitlementStore = (*Client)(nil)

func (c *Client) UpsertAppEntitlement(ctx context.Context, entitlement model.AppEntitlement) (*model.AppEntitlement, error) {
	row, err := c.Q.UpsertAppEntitlement(ctx, pgmodel.UpsertAppEntitlementParams{
		ID:    entitlement.ID,
		AppID: string(entitlement.AppID),
		UserID: pgtype.Text{
			String: string(entitlement.UserID.Value),
			Valid:  entitlement.UserID.Valid,
		},
		Source:                              string(entitlement.Source),
		SourceID:                            nullStringToText(entitlement.SourceID),
		Name:                                nullStringToText(entitlement.Name),
		Description:                         nullStringToText(entitlement.Description),
		FeatureMonthlyExecutionTimeLimit:    int32(entitlement.Features.MonthlyExecutionTimeLimit.Milliseconds()),
		FeatureMonthlyExecutionTimeAdditive: entitlement.Features.MonthlyExecutionTimeAdditive,
		CreatedAt:                           timeToTimestamp(entitlement.CreatedAt),
		UpdatedAt:                           timeToTimestamp(entitlement.UpdatedAt),
		ValidFrom:                           nullTimeToTimestamp(entitlement.ValidFrom),
		ValidUntil:                          nullTimeToTimestamp(entitlement.ValidUntil),
	})
	if err != nil {
		return nil, err
	}

	res := appEntitlementToModel(row)
	return &res, nil
}

func (c *Client) GetAppEntitlements(ctx context.Context, appID distype.Snowflake, validAt time.Time) ([]model.AppEntitlement, error) {
	rows, err := c.Q.GetAppEntitlements(ctx, pgmodel.GetAppEntitlementsParams{
		AppID:   string(appID),
		ValidAt: timeToTimestamp(validAt),
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.AppEntitlement, len(rows))
	for i, row := range rows {
		res[i] = appEntitlementToModel(row)
	}

	return res, nil
}

func (c *Client) GetResolvedAppEntitlement(ctx context.Context, appID distype.Snowflake) (*model.AppEntitlementResolved, error) {
	row, err := c.Q.GetResolvedAppEntitlement(ctx, string(appID))
	if err != nil {
		return nil, err
	}

	return &model.AppEntitlementResolved{
		MonthlyExecutionTimeLimit: time.Duration(row.FeatureMonthlyExecutionTimeLimit) * time.Millisecond,
	}, nil
}

func appEntitlementToModel(row pgmodel.AppEntitlement) model.AppEntitlement {
	return model.AppEntitlement{
		ID:    row.ID,
		AppID: distype.Snowflake(row.AppID),
		UserID: distype.Nullable[distype.Snowflake]{
			Value: distype.Snowflake(row.UserID.String),
			Valid: row.UserID.Valid,
		},
		Source:      model.AppEntitlementSource(row.Source),
		SourceID:    textToNullString(row.SourceID),
		Name:        textToNullString(row.Name),
		Description: textToNullString(row.Description),
		Features: model.AppEntitlementFeatures{
			MonthlyExecutionTimeLimit:    time.Duration(row.FeatureMonthlyExecutionTimeLimit) * time.Millisecond,
			MonthlyExecutionTimeAdditive: row.FeatureMonthlyExecutionTimeAdditive,
		},
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
		ValidFrom:  timestampToNullTime(row.ValidFrom),
		ValidUntil: timestampToNullTime(row.ValidUntil),
	}
}
