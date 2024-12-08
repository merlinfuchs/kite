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
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) EventListenersByApp(ctx context.Context, appID string) ([]*model.EventListener, error) {
	rows, err := c.Q.GetEventListenersByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	listeners := make([]*model.EventListener, len(rows))
	for i, row := range rows {
		listener, err := rowToEventListener(row)
		if err != nil {
			return nil, err
		}

		listeners[i] = listener
	}

	return listeners, nil
}

func (c *Client) CountEventListenersByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountEventListenersByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) EventListener(ctx context.Context, id string) (*model.EventListener, error) {
	row, err := c.Q.GetEventListener(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToEventListener(row)
}

func (c *Client) CreateEventListener(ctx context.Context, listener *model.EventListener) (*model.EventListener, error) {
	flowSource, err := json.Marshal(listener.FlowSource)
	if err != nil {
		return nil, err
	}

	var rawFilter []byte
	if listener.Filter != nil {
		rawFilter, err = json.Marshal(listener.Filter)
		if err != nil {
			return nil, err
		}
	}

	row, err := c.Q.CreateEventListener(ctx, pgmodel.CreateEventListenerParams{
		ID:          listener.ID,
		Source:      string(listener.Source),
		Type:        string(listener.Type),
		Description: listener.Description,
		Enabled:     listener.Enabled,
		AppID:       listener.AppID,
		ModuleID: pgtype.Text{
			String: listener.ModuleID.String,
			Valid:  listener.ModuleID.Valid,
		},
		CreatorUserID: listener.CreatorUserID,
		Filter:        rawFilter,
		FlowSource:    flowSource,
		CreatedAt:     pgtype.Timestamp{Time: listener.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: listener.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToEventListener(row)
}

func (c *Client) UpdateEventListener(ctx context.Context, listener *model.EventListener) (*model.EventListener, error) {
	flowSource, err := json.Marshal(listener.FlowSource)
	if err != nil {
		return nil, err
	}

	var rawFilter []byte
	if listener.Filter != nil {
		rawFilter, err = json.Marshal(listener.Filter)
		if err != nil {
			return nil, err
		}
	}

	row, err := c.Q.UpdateEventListener(ctx, pgmodel.UpdateEventListenerParams{
		ID:          listener.ID,
		Enabled:     listener.Enabled,
		Type:        string(listener.Type),
		Description: listener.Description,
		Filter:      rawFilter,
		FlowSource:  flowSource,
		UpdatedAt:   pgtype.Timestamp{Time: listener.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToEventListener(row)
}

func (c *Client) EnabledEventListenersUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.EventListener, error) {
	rows, err := c.Q.GetEnabledEventListenersUpdatesSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	listeners := make([]*model.EventListener, len(rows))
	for i, row := range rows {
		listener, err := rowToEventListener(row)
		if err != nil {
			return nil, err
		}

		listeners[i] = listener
	}

	return listeners, nil
}

func (c *Client) EnabledEventListenerIDs(ctx context.Context) ([]string, error) {
	return c.Q.GetEnabledEventListenerIDs(ctx)
}

func (c *Client) DeleteEventListener(ctx context.Context, id string) error {
	err := c.Q.DeleteEventListener(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}

		return err
	}

	return nil
}

func rowToEventListener(row pgmodel.EventListener) (*model.EventListener, error) {
	var flowSource flow.FlowData
	if err := json.Unmarshal(row.FlowSource, &flowSource); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flow source: %w", err)
	}

	var filter *model.EventListenerFilter
	if row.Filter != nil {
		if err := json.Unmarshal(row.Filter, &filter); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filter: %w", err)
		}
	}

	return &model.EventListener{
		ID:            row.ID,
		Source:        model.EventSource(row.Source),
		Type:          model.EventListenerType(row.Type),
		Description:   row.Description,
		Enabled:       row.Enabled,
		AppID:         row.AppID,
		ModuleID:      null.NewString(row.ModuleID.String, row.ModuleID.Valid),
		CreatorUserID: row.CreatorUserID,
		Filter:        filter,
		FlowSource:    flowSource,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}
