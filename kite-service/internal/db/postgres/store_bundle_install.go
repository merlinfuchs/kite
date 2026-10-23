package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func (c *Client) BundleInstallsByApp(ctx context.Context, appID string) ([]*model.BundleInstall, error) {
	rows, err := c.Q.GetBundleInstallsByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	installs := make([]*model.BundleInstall, len(rows))
	for i, row := range rows {
		install, err := rowToBundleInstall(row)
		if err != nil {
			return nil, err
		}
		installs[i] = install
	}

	return installs, nil
}

func (c *Client) CreateBundleInstall(ctx context.Context, install *model.BundleInstall) (*model.BundleInstall, error) {
	settings, err := json.Marshal(install.Settings)
	if err != nil {
		return nil, err
	}

	items, err := json.Marshal(install.Items)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.CreateBundleInstall(ctx, pgmodel.CreateBundleInstallParams{
		ID:            install.ID,
		Name:          install.Name,
		Description:   install.Description,
		Enabled:       install.Enabled,
		AppID:         install.AppID,
		CreatorUserID: install.CreatorUserID,
		BundleID:      install.BundleID,
		BundleVersion: int32(install.BundleVersion),
		Settings:      settings,
		Resources:     items,
		CreatedAt:     pgtype.Timestamp{Time: install.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: install.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToBundleInstall(row)
}

func rowToBundleInstall(row pgmodel.BundleInstall) (*model.BundleInstall, error) {
	var settings map[string]string
	if err := json.Unmarshal(row.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal install settings: %w", err)
	}

	var items []model.BundleInstallItem
	if err := json.Unmarshal(row.Resources, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal install resources: %w", err)
	}

	return &model.BundleInstall{
		ID:            row.ID,
		Name:          row.Name,
		Description:   row.Description,
		Enabled:       row.Enabled,
		AppID:         row.AppID,
		CreatorUserID: row.CreatorUserID,
		BundleID:      row.BundleID,
		BundleVersion: int(row.BundleVersion),
		Settings:      settings,
		Items:         items,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}
