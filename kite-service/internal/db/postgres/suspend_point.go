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

func (c *Client) CreateSuspendPoint(ctx context.Context, suspendPoint *model.SuspendPoint) (*model.SuspendPoint, error) {
	flowState, err := json.Marshal(suspendPoint.FlowState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow state: %w", err)
	}

	row, err := c.Q.CreateSuspendPoint(ctx, pgmodel.CreateSuspendPointParams{
		ID:              suspendPoint.ID,
		Type:            string(suspendPoint.Type),
		AppID:           suspendPoint.AppID,
		CommandID:       pgtype.Text{String: suspendPoint.CommandID.String, Valid: suspendPoint.CommandID.Valid},
		EventListenerID: pgtype.Text{String: suspendPoint.EventListenerID.String, Valid: suspendPoint.EventListenerID.Valid},
		MessageID:       pgtype.Text{String: suspendPoint.MessageID.String, Valid: suspendPoint.MessageID.Valid},
		FlowSourceID:    pgtype.Text{String: suspendPoint.FlowSourceID.String, Valid: suspendPoint.FlowSourceID.Valid},
		FlowNodeID:      suspendPoint.FlowNodeID,
		FlowState:       flowState,
		CreatedAt:       pgtype.Timestamp{Time: suspendPoint.CreatedAt, Valid: true},
		ExpiresAt:       pgtype.Timestamp{Time: suspendPoint.ExpiresAt.Time, Valid: suspendPoint.ExpiresAt.Valid},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create suspend point: %w", err)
	}

	return rowToSuspendPoint(row)
}

func (c *Client) DeleteSuspendPoint(ctx context.Context, id string) error {
	return c.Q.DeleteSuspendPoint(ctx, id)
}

func (c *Client) DeleteExpiredSuspendPoints(ctx context.Context, now time.Time) error {
	return c.Q.DeleteExpiredSuspendPoints(ctx, pgtype.Timestamp{Time: now, Valid: true})
}

func (c *Client) SuspendPoint(ctx context.Context, id string) (*model.SuspendPoint, error) {
	row, err := c.Q.SuspendPoint(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToSuspendPoint(row)
}

func rowToSuspendPoint(row pgmodel.SuspendPoint) (*model.SuspendPoint, error) {
	var flowState flow.FlowContextState
	err := json.Unmarshal(row.FlowState, &flowState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal flow state: %w", err)
	}

	return &model.SuspendPoint{
		ID:              row.ID,
		Type:            model.SuspendPointType(row.Type),
		AppID:           row.AppID,
		CommandID:       null.NewString(row.CommandID.String, row.CommandID.Valid),
		EventListenerID: null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
		MessageID:       null.NewString(row.MessageID.String, row.MessageID.Valid),
		FlowSourceID:    null.NewString(row.FlowSourceID.String, row.FlowSourceID.Valid),
		FlowNodeID:      row.FlowNodeID,
		FlowState:       flowState,
		CreatedAt:       row.CreatedAt.Time,
		ExpiresAt:       null.NewTime(row.ExpiresAt.Time, row.ExpiresAt.Valid),
	}, nil
}
