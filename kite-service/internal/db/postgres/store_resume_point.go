package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateResumePoint(ctx context.Context, resumePoint *model.ResumePoint) error {
	flowState, err := json.Marshal(resumePoint.FlowState)
	if err != nil {
		return fmt.Errorf("failed to marshal flow state: %w", err)
	}

	err = c.Q.CreateResumePoint(ctx, pgmodel.CreateResumePointParams{
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
		ResumeAt:          pgtype.Timestamp{Time: resumePoint.ResumeAt.Time, Valid: resumePoint.ResumeAt.Valid},
		InteractionToken:  pgtype.Text{String: resumePoint.InteractionToken.String, Valid: resumePoint.InteractionToken.Valid},
	})
	if err != nil {
		return fmt.Errorf("failed to create resume point: %w", err)
	}

	return nil
}

func (c *Client) CountPendingTimerResumePoints(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountPendingTimerResumePoints(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) HasDueTimerResumePoints(ctx context.Context, now time.Time) (bool, error) {
	return c.Q.HasDueTimerResumePoints(ctx, pgtype.Timestamp{Time: now.UTC(), Valid: true})
}

func (c *Client) LeaseDueTimerResumePoints(ctx context.Context, appIDs []string, now time.Time, leaseUntil time.Time, batchSize int) ([]*model.ResumePoint, error) {
	rows, err := c.Q.LeaseDueTimerResumePoints(ctx, pgmodel.LeaseDueTimerResumePointsParams{
		LeaseUntil: pgtype.Timestamp{Time: leaseUntil.UTC(), Valid: true},
		Now:        pgtype.Timestamp{Time: now.UTC(), Valid: true},
		AppIds:     appIDs,
		BatchSize:  int32(batchSize),
	})
	if err != nil {
		return nil, err
	}

	res := make([]*model.ResumePoint, 0, len(rows))
	for _, row := range rows {
		resumePoint, err := rowToResumePoint(row)
		if err != nil {
			// It can't be decoded on a retry either.
			slog.Error(
				"Dropping timer resume point that can't be decoded",
				slog.String("resume_point_id", row.ID),
				slog.String("error", err.Error()),
			)
			if _, err := c.Q.DeleteTimerResumePoint(ctx, pgmodel.DeleteTimerResumePointParams{ID: row.ID, AppID: row.AppID}); err != nil {
				slog.Error(
					"Failed to delete timer resume point",
					slog.String("resume_point_id", row.ID),
					slog.String("error", err.Error()),
				)
			}
			continue
		}
		res = append(res, resumePoint)
	}
	return res, nil
}

func (c *Client) DeleteTimerResumePoint(ctx context.Context, appID string, id string) (bool, error) {
	n, err := c.Q.DeleteTimerResumePoint(ctx, pgmodel.DeleteTimerResumePointParams{
		ID:    id,
		AppID: appID,
	})
	return n == 1, err
}

func (c *Client) DeleteResumePoint(ctx context.Context, id string) error {
	return c.Q.DeleteResumePoint(ctx, id)
}

func (c *Client) DeleteStaleResumePoints(ctx context.Context, now time.Time, usedBefore time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteStaleResumePoints(ctx, pgmodel.DeleteStaleResumePointsParams{
		Now:        pgtype.Timestamp{Time: now, Valid: true},
		UsedBefore: pgtype.Timestamp{Time: usedBefore, Valid: true},
		BatchSize:  int32(batchSize),
	})
}

func (c *Client) TouchResumePoint(ctx context.Context, appID string, id string, usedAt time.Time) error {
	return c.Q.TouchResumePoint(ctx, pgmodel.TouchResumePointParams{
		UsedAt: pgtype.Timestamp{Time: usedAt, Valid: true},
		ID:     id,
		AppID:  appID,
	})
}

func (c *Client) ResumePoint(ctx context.Context, appID string, id string) (*model.ResumePoint, error) {
	row, err := c.Q.ResumePoint(ctx, pgmodel.ResumePointParams{
		ID:    id,
		AppID: appID,
	})
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
		LastUsedAt:        row.LastUsedAt.Time,
		ResumeAt:          null.NewTime(row.ResumeAt.Time, row.ResumeAt.Valid),
		InteractionToken:  null.NewString(row.InteractionToken.String, row.InteractionToken.Valid),
	}, nil
}
