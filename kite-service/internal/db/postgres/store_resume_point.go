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

func (c *Client) CreateResumePoint(ctx context.Context, resumePoint *model.ResumePoint) (*model.ResumePoint, error) {
	flowState, err := json.Marshal(resumePoint.FlowState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal flow state: %w", err)
	}

	row, err := c.Q.CreateResumePoint(ctx, pgmodel.CreateResumePointParams{
		ID:                resumePoint.ID,
		Type:              string(resumePoint.Type),
		AppID:             resumePoint.AppID,
		CommandID:         pgtype.Text{String: resumePoint.CommandID.String, Valid: resumePoint.CommandID.Valid},
		EventListenerID:   pgtype.Text{String: resumePoint.EventListenerID.String, Valid: resumePoint.EventListenerID.Valid},
		MessageID:         pgtype.Text{String: resumePoint.MessageID.String, Valid: resumePoint.MessageID.Valid},
		MessageInstanceID: pgtype.Int8{Int64: resumePoint.MessageInstanceID.Int64, Valid: resumePoint.MessageInstanceID.Valid},
		FlowSourceID:      pgtype.Text{String: resumePoint.FlowSourceID.String, Valid: resumePoint.FlowSourceID.Valid},
		FlowNodeID:        resumePoint.FlowNodeID,
		FlowState:         flowState,
		CreatedAt:         pgtype.Timestamp{Time: resumePoint.CreatedAt, Valid: true},
		ExpiresAt:         pgtype.Timestamp{Time: resumePoint.ExpiresAt.Time, Valid: resumePoint.ExpiresAt.Valid},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create resume point: %w", err)
	}

	return rowToResumePoint(row)
}

func (c *Client) DeleteResumePoint(ctx context.Context, id string) error {
	return c.Q.DeleteResumePoint(ctx, id)
}

func (c *Client) DeleteExpiredResumePoints(ctx context.Context, now time.Time) error {
	return c.Q.DeleteExpiredResumePoints(ctx, pgtype.Timestamp{Time: now, Valid: true})
}

func (c *Client) ResumePoint(ctx context.Context, id string) (*model.ResumePoint, error) {
	row, err := c.Q.ResumePoint(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToResumePoint(row)
}

func rowToResumePoint(row pgmodel.ResumePoint) (*model.ResumePoint, error) {
	var flowState flow.FlowContextState
	err := json.Unmarshal(row.FlowState, &flowState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal flow state: %w", err)
	}

	return &model.ResumePoint{
		ID:                row.ID,
		Type:              model.ResumePointType(row.Type),
		AppID:             row.AppID,
		CommandID:         null.NewString(row.CommandID.String, row.CommandID.Valid),
		EventListenerID:   null.NewString(row.EventListenerID.String, row.EventListenerID.Valid),
		MessageID:         null.NewString(row.MessageID.String, row.MessageID.Valid),
		MessageInstanceID: null.NewInt(row.MessageInstanceID.Int64, row.MessageInstanceID.Valid),
		FlowSourceID:      null.NewString(row.FlowSourceID.String, row.FlowSourceID.Valid),
		FlowNodeID:        row.FlowNodeID,
		FlowState:         flowState,
		CreatedAt:         row.CreatedAt.Time,
		ExpiresAt:         null.NewTime(row.ExpiresAt.Time, row.ExpiresAt.Valid),
	}, nil
}
