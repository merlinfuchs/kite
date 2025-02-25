package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) Entitlements(ctx context.Context, appID string) ([]*model.Entitlement, error) {
	rows, err := c.Q.GetEntitlements(ctx, appID)
	if err != nil {
		return nil, err
	}

	entitlements := make([]*model.Entitlement, 0, len(rows))
	for _, row := range rows {
		entitlements = append(entitlements, rowToEntitlement(row))
	}

	return entitlements, nil
}

func (c *Client) UpsertSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error) {
	row, err := c.Q.UpsertSubscriptionEntitlement(ctx, pgmodel.UpsertSubscriptionEntitlementParams{
		ID:                    entitlement.ID,
		Type:                  entitlement.Type,
		SubscriptionID:        pgtype.Text{String: entitlement.SubscriptionID.String, Valid: entitlement.SubscriptionID.Valid},
		AppID:                 entitlement.AppID,
		LemonsqueezyProductID: pgtype.Text{String: entitlement.LemonSqueezyProductID.String, Valid: entitlement.LemonSqueezyProductID.Valid},
		LemonsqueezyVariantID: pgtype.Text{String: entitlement.LemonSqueezyVariantID.String, Valid: entitlement.LemonSqueezyVariantID.Valid},
		CreatedAt:             pgtype.Timestamp{Time: entitlement.CreatedAt, Valid: true},
		UpdatedAt:             pgtype.Timestamp{Time: entitlement.UpdatedAt, Valid: true},
		EndsAt:                pgtype.Timestamp{Time: entitlement.EndsAt.Time, Valid: entitlement.EndsAt.Valid},
	})
	if err != nil {
		return nil, err
	}

	return rowToEntitlement(row), nil
}

func (c *Client) UpdateSubscriptionEntitlement(ctx context.Context, entitlement model.Entitlement) (*model.Entitlement, error) {
	row, err := c.Q.UpdateSubscriptionEntitlement(ctx, pgmodel.UpdateSubscriptionEntitlementParams{
		SubscriptionID:        pgtype.Text{String: entitlement.SubscriptionID.String, Valid: entitlement.SubscriptionID.Valid},
		LemonsqueezyProductID: pgtype.Text{String: entitlement.LemonSqueezyProductID.String, Valid: entitlement.LemonSqueezyProductID.Valid},
		LemonsqueezyVariantID: pgtype.Text{String: entitlement.LemonSqueezyVariantID.String, Valid: entitlement.LemonSqueezyVariantID.Valid},
		UpdatedAt:             pgtype.Timestamp{Time: entitlement.UpdatedAt, Valid: true},
		EndsAt:                pgtype.Timestamp{Time: entitlement.EndsAt.Time, Valid: entitlement.EndsAt.Valid},
	})
	if err != nil {
		return nil, err
	}

	return rowToEntitlement(row), nil
}

func rowToEntitlement(row pgmodel.Entitlement) *model.Entitlement {
	return &model.Entitlement{
		ID:                    row.ID,
		Type:                  row.Type,
		SubscriptionID:        null.NewString(row.SubscriptionID.String, row.SubscriptionID.Valid),
		AppID:                 row.AppID,
		LemonSqueezyProductID: null.NewString(row.LemonsqueezyProductID.String, row.LemonsqueezyProductID.Valid),
		LemonSqueezyVariantID: null.NewString(row.LemonsqueezyVariantID.String, row.LemonsqueezyVariantID.Valid),
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
		EndsAt:                null.NewTime(row.EndsAt.Time, row.EndsAt.Valid),
	}
}
