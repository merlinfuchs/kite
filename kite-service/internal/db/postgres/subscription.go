package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) Subscriptions(ctx context.Context, userID string) ([]*model.Subscription, error) {
	rows, err := c.Q.GetSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	subs := make([]*model.Subscription, 0, len(rows))
	for _, row := range rows {
		subs = append(subs, rowToSubscription(row))
	}

	return subs, nil
}

func (c *Client) UpsertLemonSqueezySubscription(ctx context.Context, sub model.Subscription) (*model.Subscription, error) {
	row, err := c.Q.UpsertLemonSqueezySubscription(ctx, pgmodel.UpsertLemonSqueezySubscriptionParams{
		ID:                         sub.ID,
		Source:                     string(sub.Source),
		Status:                     sub.Status,
		StatusFormatted:            sub.StatusFormatted,
		RenewsAt:                   pgtype.Timestamp{Time: sub.RenewsAt, Valid: true},
		TrialEndsAt:                pgtype.Timestamp{Time: sub.TrialEndsAt.Time, Valid: sub.TrialEndsAt.Valid},
		EndsAt:                     pgtype.Timestamp{Time: sub.EndsAt.Time, Valid: sub.EndsAt.Valid},
		CreatedAt:                  pgtype.Timestamp{Time: sub.CreatedAt, Valid: true},
		UpdatedAt:                  pgtype.Timestamp{Time: sub.UpdatedAt, Valid: true},
		UserID:                     sub.UserID,
		LemonsqueezySubscriptionID: pgtype.Text{String: sub.LemonsqueezySubscriptionID.String, Valid: sub.LemonsqueezySubscriptionID.Valid},
		LemonsqueezyCustomerID:     pgtype.Text{String: sub.LemonsqueezyCustomerID.String, Valid: sub.LemonsqueezyCustomerID.Valid},
		LemonsqueezyOrderID:        pgtype.Text{String: sub.LemonsqueezyOrderID.String, Valid: sub.LemonsqueezyOrderID.Valid},
		LemonsqueezyProductID:      pgtype.Text{String: sub.LemonsqueezyProductID.String, Valid: sub.LemonsqueezyProductID.Valid},
		LemonsqueezyVariantID:      pgtype.Text{String: sub.LemonsqueezyVariantID.String, Valid: sub.LemonsqueezyVariantID.Valid},
	})
	if err != nil {
		return nil, err
	}

	return rowToSubscription(row), nil
}

func rowToSubscription(row pgmodel.Subscription) *model.Subscription {
	if row.ID == "" {
		return nil
	}

	return &model.Subscription{
		ID:                         row.ID,
		Source:                     model.SubscriptionSource(row.Source),
		Status:                     row.Status,
		StatusFormatted:            row.StatusFormatted,
		RenewsAt:                   row.RenewsAt.Time,
		TrialEndsAt:                null.NewTime(row.TrialEndsAt.Time, row.TrialEndsAt.Valid),
		EndsAt:                     null.NewTime(row.EndsAt.Time, row.EndsAt.Valid),
		CreatedAt:                  row.CreatedAt.Time,
		UpdatedAt:                  row.UpdatedAt.Time,
		UserID:                     row.UserID,
		LemonsqueezySubscriptionID: null.NewString(row.LemonsqueezySubscriptionID.String, row.LemonsqueezySubscriptionID.Valid),
		LemonsqueezyCustomerID:     null.NewString(row.LemonsqueezyCustomerID.String, row.LemonsqueezyCustomerID.Valid),
		LemonsqueezyOrderID:        null.NewString(row.LemonsqueezyOrderID.String, row.LemonsqueezyOrderID.Valid),
		LemonsqueezyProductID:      null.NewString(row.LemonsqueezyProductID.String, row.LemonsqueezyProductID.Valid),
		LemonsqueezyVariantID:      null.NewString(row.LemonsqueezyVariantID.String, row.LemonsqueezyVariantID.Valid),
	}
}
