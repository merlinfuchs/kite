package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
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
		AppID:         row.AppID,
		IntegrationID: row.IntegrationID,
		Enabled:       row.Enabled,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}
