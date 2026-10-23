package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func (c *Client) ModulesByApp(ctx context.Context, appID string) ([]*model.Module, error) {
	rows, err := c.Q.GetModulesByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	modules := make([]*model.Module, len(rows))
	for i, row := range rows {
		module, err := rowToModule(row)
		if err != nil {
			return nil, err
		}
		modules[i] = module
	}

	return modules, nil
}

func (c *Client) CreateModule(ctx context.Context, module *model.Module) (*model.Module, error) {
	settings, err := json.Marshal(module.Settings)
	if err != nil {
		return nil, err
	}

	items, err := json.Marshal(module.Items)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.CreateModule(ctx, pgmodel.CreateModuleParams{
		ID:              module.ID,
		Name:            module.Name,
		Description:     module.Description,
		Enabled:         module.Enabled,
		AppID:           module.AppID,
		CreatorUserID:   module.CreatorUserID,
		TemplateID:      module.TemplateID,
		TemplateVersion: int32(module.TemplateVersion),
		Settings:        settings,
		Resources:       items,
		CreatedAt:       pgtype.Timestamp{Time: module.CreatedAt.UTC(), Valid: true},
		UpdatedAt:       pgtype.Timestamp{Time: module.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToModule(row)
}

func rowToModule(row pgmodel.Module) (*model.Module, error) {
	var settings map[string]string
	if err := json.Unmarshal(row.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal module settings: %w", err)
	}

	var items []model.ModuleItem
	if err := json.Unmarshal(row.Resources, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal module resources: %w", err)
	}

	return &model.Module{
		ID:              row.ID,
		Name:            row.Name,
		Description:     row.Description,
		Enabled:         row.Enabled,
		AppID:           row.AppID,
		CreatorUserID:   row.CreatorUserID,
		TemplateID:      row.TemplateID,
		TemplateVersion: int(row.TemplateVersion),
		Settings:        settings,
		Items:           items,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}
