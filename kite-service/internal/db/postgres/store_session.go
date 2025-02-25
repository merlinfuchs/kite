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

func (c *Client) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	row, err := c.Q.CreateSession(ctx, pgmodel.CreateSessionParams{
		KeyHash:   session.KeyHash,
		UserID:    session.UserID,
		CreatedAt: pgtype.Timestamp{Time: session.CreatedAt.UTC(), Valid: true},
		ExpiresAt: pgtype.Timestamp{Time: session.ExpiresAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToSession(row), nil
}

func (c *Client) DeleteSession(ctx context.Context, keyHash string) error {
	return c.Q.DeleteSession(ctx, keyHash)
}

func (c *Client) Session(ctx context.Context, keyHash string) (*model.Session, error) {
	row, err := c.Q.GetSession(ctx, keyHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToSession(row), nil
}

func rowToSession(row pgmodel.Session) *model.Session {
	return &model.Session{
		KeyHash:   row.KeyHash,
		UserID:    row.UserID,
		CreatedAt: row.CreatedAt.Time,
		ExpiresAt: row.ExpiresAt.Time,
	}
}
