package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
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

func (c *Client) BundleInstall(ctx context.Context, appID string, id string) (*model.BundleInstall, error) {
	row, err := c.Q.GetBundleInstall(ctx, pgmodel.GetBundleInstallParams{ID: id, AppID: appID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToBundleInstall(row)
}

func (c *Client) CreateBundleInstall(ctx context.Context, install *model.BundleInstall, contents store.BundleInstallContents) (*model.BundleInstall, error) {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	settings, err := json.Marshal(install.Settings)
	if err != nil {
		return nil, err
	}

	items, err := json.Marshal(install.Items)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.WithTx(tx).CreateBundleInstall(ctx, pgmodel.CreateBundleInstallParams{
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

	if err := c.createBundleInstallContentsWithTx(ctx, tx, contents); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rowToBundleInstall(row)
}

func (c *Client) createBundleInstallContentsWithTx(ctx context.Context, tx pgx.Tx, contents store.BundleInstallContents) (err error) {
	for i, variable := range contents.Variables {
		if contents.Variables[i], err = c.createVariableWithTx(ctx, tx, variable); err != nil {
			return fmt.Errorf("failed to create variable: %w", err)
		}
	}
	for i, msg := range contents.Messages {
		if contents.Messages[i], err = c.createMessageWithTx(ctx, tx, msg); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
	}
	for i, command := range contents.Commands {
		if contents.Commands[i], err = c.createCommandWithTx(ctx, tx, command); err != nil {
			return fmt.Errorf("failed to create command: %w", err)
		}
	}
	for i, listener := range contents.EventListeners {
		if contents.EventListeners[i], err = c.createEventListenerWithTx(ctx, tx, listener); err != nil {
			return fmt.Errorf("failed to create event listener: %w", err)
		}
	}
	return nil
}

func (c *Client) UpdateBundleInstall(ctx context.Context, install *model.BundleInstall, changes store.BundleInstallChanges) (*model.BundleInstall, error) {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	settings, err := json.Marshal(install.Settings)
	if err != nil {
		return nil, err
	}

	items, err := json.Marshal(install.Items)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.WithTx(tx).UpdateBundleInstall(ctx, pgmodel.UpdateBundleInstallParams{
		ID:            install.ID,
		BundleVersion: int32(install.BundleVersion),
		Settings:      settings,
		Resources:     items,
		UpdatedAt:     pgtype.Timestamp{Time: install.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	if err := c.createBundleInstallContentsWithTx(ctx, tx, changes.Created); err != nil {
		return nil, err
	}

	contents := changes.Updated
	for i, msg := range contents.Messages {
		if contents.Messages[i], err = c.updateMessageWithTx(ctx, tx, msg); err != nil {
			return nil, fmt.Errorf("failed to update message: %w", err)
		}
	}
	for i, command := range contents.Commands {
		if contents.Commands[i], err = c.updateCommandWithTx(ctx, tx, command); err != nil {
			return nil, fmt.Errorf("failed to update command: %w", err)
		}
	}
	for i, listener := range contents.EventListeners {
		if contents.EventListeners[i], err = c.updateEventListenerWithTx(ctx, tx, listener); err != nil {
			return nil, fmt.Errorf("failed to update event listener: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rowToBundleInstall(row)
}

func (c *Client) DeleteBundleInstall(ctx context.Context, id string) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	if err := q.DeleteCommandsByBundleInstall(ctx, pgtype.Text{String: id, Valid: true}); err != nil {
		return fmt.Errorf("failed to delete commands: %w", err)
	}
	if err := q.DeleteEventListenersByBundleInstall(ctx, pgtype.Text{String: id, Valid: true}); err != nil {
		return fmt.Errorf("failed to delete event listeners: %w", err)
	}
	if err := q.DeleteMessagesByBundleInstall(ctx, pgtype.Text{String: id, Valid: true}); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	if err := q.DeleteVariablesByBundleInstall(ctx, pgtype.Text{String: id, Valid: true}); err != nil {
		return fmt.Errorf("failed to delete variables: %w", err)
	}
	if err := q.DeleteBundleInstall(ctx, id); err != nil {
		return fmt.Errorf("failed to delete install: %w", err)
	}

	return tx.Commit(ctx)
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
