package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CommandsByApp(ctx context.Context, appID string) ([]*model.Command, error) {
	rows, err := c.Q.GetCommandsByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	commands := make([]*model.Command, len(rows))
	for i, row := range rows {
		cmd, err := rowToCommand(row)
		if err != nil {
			return nil, err
		}

		commands[i] = cmd
	}

	return commands, nil
}

func (c *Client) CountCommandsByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountCommandsByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) Command(ctx context.Context, id string) (*model.Command, error) {
	row, err := c.Q.GetCommand(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToCommand(row)
}

func (c *Client) CreateCommand(ctx context.Context, command *model.Command) (*model.Command, error) {
	flowSource, err := json.Marshal(command.FlowSource)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.CreateCommand(ctx, pgmodel.CreateCommandParams{
		ID:          command.ID,
		Name:        command.Name,
		Description: command.Description,
		Enabled:     command.Enabled,
		AppID:       command.AppID,
		ModuleID: pgtype.Text{
			String: command.ModuleID.String,
			Valid:  command.ModuleID.Valid,
		},
		CreatorUserID: command.CreatorUserID,
		FlowSource:    flowSource,
		CreatedAt:     pgtype.Timestamp{Time: command.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: command.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToCommand(row)
}

func (c *Client) UpdateCommand(ctx context.Context, command *model.Command) (*model.Command, error) {
	flowSource, err := json.Marshal(command.FlowSource)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.UpdateCommand(ctx, pgmodel.UpdateCommandParams{
		ID:          command.ID,
		Name:        command.Name,
		Description: command.Description,
		Enabled:     command.Enabled,
		FlowSource:  flowSource,
		UpdatedAt:   pgtype.Timestamp{Time: command.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToCommand(row)
}

func (c *Client) UpdateCommandsLastDeployedAt(ctx context.Context, appID string, lastDeployedAt time.Time) error {
	return c.Q.UpdateCommandsLastDeployedAt(ctx, pgmodel.UpdateCommandsLastDeployedAtParams{
		AppID:          appID,
		LastDeployedAt: pgtype.Timestamp{Time: lastDeployedAt.UTC(), Valid: true},
	})
}

func (c *Client) EnabledCommandsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.Command, error) {
	rows, err := c.Q.GetEnabledCommandsUpdatesSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	commands := make([]*model.Command, len(rows))
	for i, row := range rows {
		cmd, err := rowToCommand(row)
		if err != nil {
			return nil, err
		}

		commands[i] = cmd
	}

	return commands, nil
}

func (c *Client) EnabledCommandIDs(ctx context.Context) ([]string, error) {
	return c.Q.GetEnabledCommandIDs(ctx)
}

func (c *Client) DeleteCommand(ctx context.Context, id string) error {
	err := c.Q.DeleteCommand(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}

		return err
	}

	return nil
}

func rowToCommand(row pgmodel.Command) (*model.Command, error) {
	var flowSource flow.FlowData
	if err := json.Unmarshal(row.FlowSource, &flowSource); err != nil {
		return nil, err
	}

	return &model.Command{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		Enabled:        row.Enabled,
		AppID:          row.AppID,
		ModuleID:       null.NewString(row.ModuleID.String, row.ModuleID.Valid),
		CreatorUserID:  row.CreatorUserID,
		FlowSource:     flowSource,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		LastDeployedAt: null.NewTime(row.LastDeployedAt.Time, row.LastDeployedAt.Valid),
	}, nil
}
