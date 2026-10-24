package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateFlowVersion(ctx context.Context, version *model.FlowVersion) error {
	flowSource, err := json.Marshal(version.FlowSource)
	if err != nil {
		return err
	}

	return c.Q.CreateFlowVersion(ctx, pgmodel.CreateFlowVersionParams{
		ID:              version.ID,
		AppID:           version.AppID,
		CommandID:       pgtype.Text{String: version.CommandID.String, Valid: version.CommandID.Valid},
		EventListenerID: pgtype.Text{String: version.EventListenerID.String, Valid: version.EventListenerID.Valid},
		FlowSource:      flowSource,
		AutoSaved:       version.AutoSaved,
		CreatorUserID:   pgtype.Text{String: version.CreatorUserID.String, Valid: version.CreatorUserID.Valid},
		CreatedAt:       pgtype.Timestamp{Time: version.CreatedAt.UTC(), Valid: true},
	})
}

func (c *Client) FlowVersion(ctx context.Context, appID string, id string) (*model.FlowVersion, error) {
	row, err := c.Q.GetFlowVersion(ctx, pgmodel.GetFlowVersionParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var flowSource flow.FlowData
	if err := json.Unmarshal(row.FlowSource, &flowSource); err != nil {
		return nil, err
	}

	return &model.FlowVersion{
		ID:              row.ID,
		AppID:           row.AppID,
		CommandID:       null.NewString(row.CommandID.String, row.CommandID.Valid),
		EventListenerID: null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
		FlowSource:      flowSource,
		AutoSaved:       row.AutoSaved,
		CreatorUserID:   null.NewString(row.CreatorUserID.String, row.CreatorUserID.Valid),
		CreatedAt:       row.CreatedAt.Time,
	}, nil
}

func (c *Client) FlowVersionsByCommand(ctx context.Context, appID string, commandID string, limit int) ([]*model.FlowVersion, error) {
	rows, err := c.Q.GetFlowVersionsByCommand(ctx, pgmodel.GetFlowVersionsByCommandParams{
		AppID:     appID,
		CommandID: pgtype.Text{String: commandID, Valid: true},
		MaxCount:  int32(limit),
	})
	if err != nil {
		return nil, err
	}

	versions := make([]*model.FlowVersion, len(rows))
	for i, row := range rows {
		versions[i] = &model.FlowVersion{
			ID:                 row.ID,
			AppID:              row.AppID,
			CommandID:          null.NewString(row.CommandID.String, row.CommandID.Valid),
			EventListenerID:    null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
			AutoSaved:          row.AutoSaved,
			CreatorUserID:      null.NewString(row.CreatorUserID.String, row.CreatorUserID.Valid),
			CreatorDisplayName: null.NewString(row.CreatorDisplayName.String, row.CreatorDisplayName.Valid),
			CreatedAt:          row.CreatedAt.Time,
		}
	}

	return versions, nil
}

func (c *Client) FlowVersionsByEventListener(ctx context.Context, appID string, eventListenerID string, limit int) ([]*model.FlowVersion, error) {
	rows, err := c.Q.GetFlowVersionsByEventListener(ctx, pgmodel.GetFlowVersionsByEventListenerParams{
		AppID:           appID,
		EventListenerID: pgtype.Text{String: eventListenerID, Valid: true},
		MaxCount:        int32(limit),
	})
	if err != nil {
		return nil, err
	}

	versions := make([]*model.FlowVersion, len(rows))
	for i, row := range rows {
		versions[i] = &model.FlowVersion{
			ID:                 row.ID,
			AppID:              row.AppID,
			CommandID:          null.NewString(row.CommandID.String, row.CommandID.Valid),
			EventListenerID:    null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
			AutoSaved:          row.AutoSaved,
			CreatorUserID:      null.NewString(row.CreatorUserID.String, row.CreatorUserID.Valid),
			CreatorDisplayName: null.NewString(row.CreatorDisplayName.String, row.CreatorDisplayName.Valid),
			CreatedAt:          row.CreatedAt.Time,
		}
	}

	return versions, nil
}

func (c *Client) CountFlowVersionsByCommand(ctx context.Context, commandID string) (int, error) {
	res, err := c.Q.CountFlowVersionsByCommand(ctx, pgtype.Text{String: commandID, Valid: true})
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) CountFlowVersionsByEventListener(ctx context.Context, eventListenerID string) (int, error) {
	res, err := c.Q.CountFlowVersionsByEventListener(ctx, pgtype.Text{String: eventListenerID, Valid: true})
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) DeleteOldFlowVersionsByCommand(ctx context.Context, commandID string, keep int) error {
	return c.Q.DeleteOldFlowVersionsByCommand(ctx, pgmodel.DeleteOldFlowVersionsByCommandParams{
		CommandID: pgtype.Text{String: commandID, Valid: true},
		KeepCount: int32(keep),
	})
}

func (c *Client) DeleteOldFlowVersionsByEventListener(ctx context.Context, eventListenerID string, keep int) error {
	return c.Q.DeleteOldFlowVersionsByEventListener(ctx, pgmodel.DeleteOldFlowVersionsByEventListenerParams{
		EventListenerID: pgtype.Text{String: eventListenerID, Valid: true},
		KeepCount:       int32(keep),
	})
}
