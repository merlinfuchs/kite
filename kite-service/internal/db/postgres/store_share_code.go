package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateShareCode(ctx context.Context, shareCode *model.ShareCode) error {
	return c.Q.CreateShareCode(ctx, pgmodel.CreateShareCodeParams{
		Code:          shareCode.Code,
		Type:          string(shareCode.Type),
		Data:          shareCode.Data,
		CreatorUserID: shareCode.CreatorUserID,
		AppID:         pgtype.Text{String: shareCode.AppID.String, Valid: shareCode.AppID.Valid},
		CreatedAt:     pgtype.Timestamp{Time: shareCode.CreatedAt.UTC(), Valid: true},
		LastUsedAt:    pgtype.Timestamp{Time: shareCode.LastUsedAt.UTC(), Valid: true},
	})
}

func (c *Client) ShareCode(ctx context.Context, code string) (*model.ShareCode, error) {
	row, err := c.Q.ShareCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToShareCode(row), nil
}

func (c *Client) TouchShareCode(ctx context.Context, code string, usedAt time.Time) error {
	return c.Q.TouchShareCode(ctx, pgmodel.TouchShareCodeParams{
		UsedAt: pgtype.Timestamp{Time: usedAt.UTC(), Valid: true},
		Code:   code,
	})
}

func (c *Client) DeleteUnusedShareCodes(ctx context.Context, usedBefore time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteUnusedShareCodes(ctx, pgmodel.DeleteUnusedShareCodesParams{
		UsedBefore: pgtype.Timestamp{Time: usedBefore.UTC(), Valid: true},
		BatchSize:  int32(batchSize),
	})
}

func rowToShareCode(row pgmodel.ShareCode) *model.ShareCode {
	return &model.ShareCode{
		Code:          row.Code,
		Type:          model.ShareCodeType(row.Type),
		Data:          row.Data,
		CreatorUserID: row.CreatorUserID,
		AppID:         null.NewString(row.AppID.String, row.AppID.Valid),
		CreatedAt:     row.CreatedAt.Time,
		LastUsedAt:    row.LastUsedAt.Time,
	}
}
