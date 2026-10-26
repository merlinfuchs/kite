package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

	params, err := json.Marshal(install.Parameters)
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
		// The column is named after what parameters were called before.
		Settings:  params,
		Resources: items,
		CreatedAt: pgtype.Timestamp{Time: install.CreatedAt.UTC(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: install.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, uniqueViolation(err)
	}

	if err := c.createBundleInstallContentsWithTx(ctx, tx, contents); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rowToBundleInstall(row)
}

func (c *Client) ImportBundleContents(ctx context.Context, contents store.BundleInstallContents) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := c.createBundleInstallContentsWithTx(ctx, tx, contents); err != nil {
		return err
	}

	return tx.Commit(ctx)
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

func (c *Client) UpdateBundleInstall(ctx context.Context, install *model.BundleInstall, fromUpdatedAt time.Time, changes store.BundleInstallChanges, userID string) (*model.BundleInstall, error) {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	params, err := json.Marshal(install.Parameters)
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
		// The column is named after what parameters were called before.
		Settings:      params,
		Resources:     items,
		UpdatedAt:     pgtype.Timestamp{Time: install.UpdatedAt.UTC(), Valid: true},
		FromUpdatedAt: pgtype.Timestamp{Time: fromUpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrOutdated
		}
		return nil, err
	}

	if err := c.createBundleInstallContentsWithTx(ctx, tx, changes.Created); err != nil {
		return nil, err
	}

	contents := changes.Updated
	change := store.ItemChange{Source: model.ItemVersionSourceBundle, UserID: userID}
	for i, msg := range contents.Messages {
		if contents.Messages[i], err = c.updateMessageWithTx(ctx, tx, msg, change); err != nil {
			return nil, fmt.Errorf("failed to update message: %w", err)
		}
	}
	for i, command := range contents.Commands {
		if contents.Commands[i], err = c.updateCommandWithTx(ctx, tx, command, change); err != nil {
			return nil, fmt.Errorf("failed to update command: %w", err)
		}
	}
	for i, listener := range contents.EventListeners {
		if contents.EventListeners[i], err = c.updateEventListenerWithTx(ctx, tx, listener, change); err != nil {
			return nil, fmt.Errorf("failed to update event listener: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rowToBundleInstall(row)
}

func (c *Client) DeleteBundleInstall(ctx context.Context, id string, keepIDs []string, userID string) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// A nil slice would be NULL, which = ANY never matches, so nothing would
	// be deleted.
	if keepIDs == nil {
		keepIDs = []string{}
	}

	// The deleted commands, event listeners and messages stay in their
	// version history.
	q := c.Q.WithTx(tx)
	commands, err := q.DeleteCommandsByBundleInstall(ctx, pgtype.Text{String: id, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete commands: %w", err)
	}
	for _, row := range commands {
		if err := recordCommandDelete(ctx, q, row, userID); err != nil {
			return err
		}
	}
	listeners, err := q.DeleteEventListenersByBundleInstall(ctx, pgtype.Text{String: id, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete event listeners: %w", err)
	}
	for _, row := range listeners {
		if err := recordEventListenerDelete(ctx, q, row, userID); err != nil {
			return err
		}
	}
	messages, err := q.DeleteMessagesByBundleInstall(ctx, pgmodel.DeleteMessagesByBundleInstallParams{
		BundleInstallID: pgtype.Text{String: id, Valid: true},
		KeepIds:         keepIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	for _, row := range messages {
		if err := recordMessageDelete(ctx, q, row, userID); err != nil {
			return err
		}
	}
	err = q.DeleteVariablesByBundleInstall(ctx, pgmodel.DeleteVariablesByBundleInstallParams{
		BundleInstallID: pgtype.Text{String: id, Valid: true},
		KeepIds:         keepIDs,
	})
	if err != nil {
		return fmt.Errorf("failed to delete variables: %w", err)
	}
	if err := q.DeleteBundleInstall(ctx, id); err != nil {
		return fmt.Errorf("failed to delete install: %w", err)
	}

	return tx.Commit(ctx)
}

func rowToBundleInstall(row pgmodel.BundleInstall) (*model.BundleInstall, error) {
	var params map[string]string
	if err := json.Unmarshal(row.Settings, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal install parameters: %w", err)
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
		Parameters:    params,
		Items:         items,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (c *Client) InstalledContentsByApp(ctx context.Context, appID string) (store.BundleInstallContents, error) {
	var res store.BundleInstallContents

	commands, err := c.Q.GetInstalledCommandsByApp(ctx, appID)
	if err != nil {
		return res, fmt.Errorf("failed to get commands: %w", err)
	}
	if res.Commands, err = decodeAll(commands, rowToCommand); err != nil {
		return res, fmt.Errorf("failed to decode commands: %w", err)
	}

	listeners, err := c.Q.GetInstalledEventListenersByApp(ctx, appID)
	if err != nil {
		return res, fmt.Errorf("failed to get event listeners: %w", err)
	}
	if res.EventListeners, err = decodeAll(listeners, rowToEventListener); err != nil {
		return res, fmt.Errorf("failed to decode event listeners: %w", err)
	}

	messages, err := c.Q.GetInstalledMessagesByApp(ctx, appID)
	if err != nil {
		return res, fmt.Errorf("failed to get messages: %w", err)
	}
	if res.Messages, err = decodeAll(messages, rowToMessage); err != nil {
		return res, fmt.Errorf("failed to decode messages: %w", err)
	}

	variables, err := c.Q.GetInstalledVariablesByApp(ctx, appID)
	if err != nil {
		return res, fmt.Errorf("failed to get variables: %w", err)
	}
	res.Variables = make([]*model.Variable, len(variables))
	for i, row := range variables {
		res.Variables[i] = rowToVariable(row)
	}

	return res, nil
}
