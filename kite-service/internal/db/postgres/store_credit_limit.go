package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreditLimitsByApp(ctx context.Context, appID string) ([]*model.CreditLimit, error) {
	rows, err := c.Q.GetCreditLimitsByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	res := make([]*model.CreditLimit, len(rows))
	for i, row := range rows {
		res[i] = rowToCreditLimit(row)
	}
	return res, nil
}

func (c *Client) CreditLimit(ctx context.Context, appID string, id string) (*model.CreditLimit, error) {
	row, err := c.Q.GetCreditLimit(ctx, pgmodel.GetCreditLimitParams{AppID: appID, ID: id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return rowToCreditLimit(row), nil
}

func (c *Client) CountCreditLimitsByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountCreditLimitsByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) CreateCreditLimit(ctx context.Context, limit *model.CreditLimit) (*model.CreditLimit, error) {
	row, err := c.Q.CreateCreditLimit(ctx, pgmodel.CreateCreditLimitParams{
		ID:        limit.ID,
		AppID:     limit.AppID,
		Scope:     string(limit.Scope),
		TargetID:  pgtype.Text{String: limit.TargetID.String, Valid: limit.TargetID.Valid},
		Period:    string(limit.Period),
		Credits:   pgtype.Int4{Int32: int32(limit.Credits.Int64), Valid: limit.Credits.Valid},
		CreatedAt: pgtype.Timestamp{Time: limit.CreatedAt.UTC(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: limit.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, creditLimitWriteError(err)
	}
	return rowToCreditLimit(row), nil
}

func (c *Client) UpdateCreditLimit(ctx context.Context, limit *model.CreditLimit) (*model.CreditLimit, error) {
	row, err := c.Q.UpdateCreditLimit(ctx, pgmodel.UpdateCreditLimitParams{
		AppID:     limit.AppID,
		ID:        limit.ID,
		Scope:     string(limit.Scope),
		TargetID:  pgtype.Text{String: limit.TargetID.String, Valid: limit.TargetID.Valid},
		Period:    string(limit.Period),
		Credits:   pgtype.Int4{Int32: int32(limit.Credits.Int64), Valid: limit.Credits.Valid},
		UpdatedAt: pgtype.Timestamp{Time: limit.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, creditLimitWriteError(err)
	}
	return rowToCreditLimit(row), nil
}

func (c *Client) DeleteCreditLimit(ctx context.Context, appID string, id string) error {
	n, err := c.Q.DeleteCreditLimit(ctx, pgmodel.DeleteCreditLimitParams{AppID: appID, ID: id})
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

// creditLimitWriteError reports a limit for the same scope, target and period
// the app has already.
func creditLimitWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return store.ErrAlreadyExists
	}
	return err
}

func rowToCreditLimit(row pgmodel.CreditLimit) *model.CreditLimit {
	return &model.CreditLimit{
		ID:        row.ID,
		AppID:     row.AppID,
		Scope:     model.CreditLimitScope(row.Scope),
		TargetID:  null.NewString(row.TargetID.String, row.TargetID.Valid),
		Period:    model.CreditLimitPeriod(row.Period),
		Credits:   null.NewInt(int64(row.Credits.Int32), row.Credits.Valid),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
