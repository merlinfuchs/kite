package postgres

import (
	"context"
	"database/sql"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func (c *Client) GetSession(ctx context.Context, tokenHash string) (*model.Session, error) {
	row, err := c.Q.GetSession(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.Session{
		TokenHash:   row.TokenHash,
		UserID:      row.UserID,
		GuildIds:    row.GuildIds,
		AccessToken: row.AccessToken,
		CreatedAt:   row.CreatedAt,
		ExpiresAt:   row.ExpiresAt,
	}, nil
}

func (c *Client) DeleteSession(ctx context.Context, tokenHash string) error {
	err := c.Q.DeleteSession(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}
	return nil
}

func (c *Client) CreateSession(ctx context.Context, session *model.Session) error {
	err := c.Q.CreateSession(ctx, pgmodel.CreateSessionParams{
		TokenHash:   session.TokenHash,
		UserID:      session.UserID,
		GuildIds:    session.GuildIds,
		AccessToken: session.AccessToken,
		CreatedAt:   session.CreatedAt,
		ExpiresAt:   session.ExpiresAt,
	})
	if err != nil {
		return err
	}
	return nil
}
