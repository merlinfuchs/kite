package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (c *Client) AppIntegrations(ctx context.Context, appID string) ([]*model.AppIntegration, error) {
	rows, err := c.Q.GetAppIntegrations(ctx, appID)
	if err != nil {
		return nil, err
	}

	res := make([]*model.AppIntegration, len(rows))
	for i, row := range rows {
		res[i] = rowToAppIntegration(row)
	}
	return res, nil
}

func (c *Client) SetAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error) {
	row, err := c.Q.SetAppIntegrationEnabled(ctx, pgmodel.SetAppIntegrationEnabledParams{
		ID:            integration.ID,
		AppID:         integration.AppID,
		IntegrationID: integration.IntegrationID,
		Enabled:       integration.Enabled,
		CreatedAt:     pgtype.Timestamp{Time: integration.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: integration.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return rowToAppIntegration(row), nil
}

func rowToAppIntegration(row pgmodel.AppIntegration) *model.AppIntegration {
	return &model.AppIntegration{
		ID:            row.ID,
		AppID:         row.AppID,
		IntegrationID: row.IntegrationID,
		Enabled:       row.Enabled,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func (c *Client) UpdateAppIntegrationEnabled(ctx context.Context, integration *model.AppIntegration) (*model.AppIntegration, error) {
	row, err := c.Q.UpdateAppIntegrationEnabled(ctx, pgmodel.UpdateAppIntegrationEnabledParams{
		AppID:         integration.AppID,
		IntegrationID: integration.IntegrationID,
		Enabled:       integration.Enabled,
		UpdatedAt:     pgtype.Timestamp{Time: integration.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return rowToAppIntegration(row), nil
}

func (c *Client) ConnectAppIntegration(ctx context.Context, integration *model.AppIntegration, secret *model.AppSecret) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	q := c.Q.WithTx(tx)

	// The credential references the integration's row, which keeps its ID if
	// the app set it up before.
	row, err := q.EnsureAppIntegration(ctx, pgmodel.EnsureAppIntegrationParams{
		ID:            integration.ID,
		AppID:         integration.AppID,
		IntegrationID: integration.IntegrationID,
		CreatedAt:     pgtype.Timestamp{Time: integration.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: integration.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return err
	}

	_, err = q.SetAppIntegrationCredential(ctx, pgmodel.SetAppIntegrationCredentialParams{
		ID:               secret.ID,
		AppID:            secret.AppID,
		AppIntegrationID: pgtype.Text{String: row.ID, Valid: true},
		ValueEncrypted:   secret.ValueEncrypted,
		CreatedAt:        pgtype.Timestamp{Time: secret.CreatedAt.UTC(), Valid: true},
		UpdatedAt:        pgtype.Timestamp{Time: secret.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (c *Client) DeleteAppIntegration(ctx context.Context, appID string, integrationID string) error {
	n, err := c.Q.DeleteAppIntegration(ctx, pgmodel.DeleteAppIntegrationParams{
		AppID:         appID,
		IntegrationID: integrationID,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}
