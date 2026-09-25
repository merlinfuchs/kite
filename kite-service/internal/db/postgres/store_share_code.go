package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (c *Client) CreateShareCode(ctx context.Context, shareCode *model.ShareCode) (*model.ShareCode, error) {
	row, err := c.Q.CreateShareCode(ctx, pgmodel.CreateShareCodeParams{
		Code:      shareCode.Code,
		Data:      shareCode.Data,
		CreatedAt: pgtype.Timestamp{Time: shareCode.CreatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToShareCode(row), nil
}

func (c *Client) ShareCode(ctx context.Context, code string) (*model.ShareCode, error) {
	row, err := c.Q.GetShareCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToShareCode(row), nil
}

func rowToShareCode(row pgmodel.ShareCode) *model.ShareCode {
	return &model.ShareCode{
		Code:      row.Code,
		Data:      row.Data,
		CreatedAt: row.CreatedAt.Time,
	}
}
