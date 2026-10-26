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

func (c *Client) CreateAPIToken(ctx context.Context, token *model.APIToken) (*model.APIToken, error) {
	row, err := c.Q.CreateAPIToken(ctx, pgmodel.CreateAPITokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		Name:      token.Name,
		KeyHash:   token.KeyHash,
		Scopes:    token.Scopes,
		ExpiresAt: pgtype.Timestamp{Time: token.ExpiresAt.Time.UTC(), Valid: token.ExpiresAt.Valid},
		CreatedAt: pgtype.Timestamp{Time: token.CreatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToAPIToken(row), nil
}

func (c *Client) APITokensByUser(ctx context.Context, userID string) ([]*model.APIToken, error) {
	rows, err := c.Q.GetAPITokensByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokens := make([]*model.APIToken, len(rows))
	for i, row := range rows {
		tokens[i] = rowToAPIToken(row)
	}
	return tokens, nil
}

func (c *Client) CountAPITokensByUser(ctx context.Context, userID string, now time.Time) (int, error) {
	count, err := c.Q.CountAPITokensByUser(ctx, pgmodel.CountAPITokensByUserParams{
		UserID: userID,
		Now:    pgtype.Timestamp{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (c *Client) APITokenByKeyHash(ctx context.Context, keyHash string) (*model.APIToken, error) {
	row, err := c.Q.GetAPITokenByKeyHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToAPIToken(row), nil
}

func (c *Client) UpdateAPITokenLastUsedAt(ctx context.Context, id string, lastUsedAt time.Time, interval time.Duration) error {
	return c.Q.UpdateAPITokenLastUsedAt(ctx, pgmodel.UpdateAPITokenLastUsedAtParams{
		ID:             id,
		LastUsedAt:     pgtype.Timestamp{Time: lastUsedAt.UTC(), Valid: true},
		LastUsedBefore: pgtype.Timestamp{Time: lastUsedAt.Add(-interval).UTC(), Valid: true},
	})
}

func (c *Client) DeleteAPIToken(ctx context.Context, userID string, id string) error {
	rows, err := c.Q.DeleteAPIToken(ctx, pgmodel.DeleteAPITokenParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func rowToAPIToken(row pgmodel.ApiToken) *model.APIToken {
	return &model.APIToken{
		ID:         row.ID,
		UserID:     row.UserID,
		Name:       row.Name,
		KeyHash:    row.KeyHash,
		Scopes:     row.Scopes,
		LastUsedAt: null.NewTime(row.LastUsedAt.Time, row.LastUsedAt.Valid),
		ExpiresAt:  null.NewTime(row.ExpiresAt.Time, row.ExpiresAt.Valid),
		CreatedAt:  row.CreatedAt.Time,
	}
}
