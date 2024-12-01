package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateAPIKey(ctx context.Context, key *model.APIKey) (*model.APIKey, error) {
	row, err := c.Q.CreateAPIKey(ctx, pgmodel.CreateAPIKeyParams{
		ID:            key.ID,
		Type:          string(key.Type),
		Name:          key.Name,
		Key:           key.Key,
		KeyHash:       key.KeyHash,
		AppID:         key.AppID,
		CreatorUserID: key.CreatorUserID,
		CreatedAt:     pgtype.Timestamp{Time: key.CreatedAt, Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: key.UpdatedAt, Valid: true},
		ExpiresAt:     pgtype.Timestamp{Time: key.ExpiresAt.Time, Valid: key.ExpiresAt.Valid},
	})
	if err != nil {
		return nil, err
	}

	return rowToAPIKey(row), nil
}

func (c *Client) APIKeyByKeyHash(ctx context.Context, hash string) (*model.APIKey, error) {
	row, err := c.Q.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	return rowToAPIKey(row), nil
}

func (c *Client) APIKeysByApp(ctx context.Context, appID string) ([]*model.APIKey, error) {
	rows, err := c.Q.GetAPIKeysByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	apiKeys := make([]*model.APIKey, len(rows))
	for i, row := range rows {
		apiKeys[i] = rowToAPIKey(row)
	}

	return apiKeys, nil
}

func rowToAPIKey(row pgmodel.ApiKey) *model.APIKey {
	return &model.APIKey{
		ID:            row.ID,
		Type:          model.APIKeyType(row.Type),
		Name:          row.Name,
		Key:           row.Key,
		KeyHash:       row.KeyHash,
		AppID:         row.AppID,
		CreatorUserID: row.CreatorUserID,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		ExpiresAt:     null.NewTime(row.ExpiresAt.Time, row.ExpiresAt.Valid),
	}
}
