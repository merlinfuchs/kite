package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (c *Client) CreateFlowAIPrompt(ctx context.Context, prompt *model.FlowAIPrompt) error {
	return c.Q.CreateFlowAIPrompt(ctx, pgmodel.CreateFlowAIPromptParams{
		ID:                prompt.ID,
		AppID:             prompt.AppID,
		UserID:            prompt.UserID,
		Model:             prompt.Model,
		Rounds:            int32(prompt.Rounds),
		InputTokens:       int32(prompt.Usage.InputTokens),
		CachedInputTokens: int32(prompt.Usage.CachedInputTokens),
		OutputTokens:      int32(prompt.Usage.OutputTokens),
		CreatedAt:         pgtype.Timestamp{Time: prompt.CreatedAt, Valid: true},
		UpdatedAt:         pgtype.Timestamp{Time: prompt.UpdatedAt, Valid: true},
	})
}

func (c *Client) FlowAIPrompt(ctx context.Context, appID string, id string) (*model.FlowAIPrompt, error) {
	row, err := c.Q.GetFlowAIPrompt(ctx, pgmodel.GetFlowAIPromptParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToFlowAIPrompt(row), nil
}

func (c *Client) DeleteFlowAIPrompt(ctx context.Context, appID string, id string) error {
	return c.Q.DeleteFlowAIPrompt(ctx, pgmodel.DeleteFlowAIPromptParams{
		ID:    id,
		AppID: appID,
	})
}

func (c *Client) StartFlowAIPromptRound(ctx context.Context, appID string, id string, maxRounds int, updatedAt time.Time) (bool, error) {
	rows, err := c.Q.StartFlowAIPromptRound(ctx, pgmodel.StartFlowAIPromptRoundParams{
		ID:        id,
		AppID:     appID,
		MaxRounds: int32(maxRounds),
		UpdatedAt: pgtype.Timestamp{Time: updatedAt, Valid: true},
	})
	return rows > 0, err
}

func (c *Client) AddFlowAIPromptUsage(ctx context.Context, appID string, id string, usage model.FlowAIUsage, updatedAt time.Time) error {
	return c.Q.AddFlowAIPromptUsage(ctx, pgmodel.AddFlowAIPromptUsageParams{
		ID:                id,
		AppID:             appID,
		InputTokens:       int32(usage.InputTokens),
		CachedInputTokens: int32(usage.CachedInputTokens),
		OutputTokens:      int32(usage.OutputTokens),
		UpdatedAt:         pgtype.Timestamp{Time: updatedAt, Valid: true},
	})
}

func (c *Client) CountFlowAIPromptsBetween(ctx context.Context, appID string, start time.Time, end time.Time) (int, error) {
	count, err := c.Q.CountFlowAIPromptsByAppBetween(ctx, pgmodel.CountFlowAIPromptsByAppBetweenParams{
		AppID:   appID,
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	return int(count), err
}

func rowToFlowAIPrompt(row pgmodel.FlowAiPrompt) *model.FlowAIPrompt {
	return &model.FlowAIPrompt{
		ID:     row.ID,
		AppID:  row.AppID,
		UserID: row.UserID,
		Model:  row.Model,
		Rounds: int(row.Rounds),
		Usage: model.FlowAIUsage{
			InputTokens:       int(row.InputTokens),
			CachedInputTokens: int(row.CachedInputTokens),
			OutputTokens:      int(row.OutputTokens),
		},
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
