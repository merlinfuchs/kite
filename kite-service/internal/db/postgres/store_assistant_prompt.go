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

func (c *Client) CreateAssistantPrompt(ctx context.Context, prompt *model.AssistantPrompt) error {
	return c.Q.CreateAssistantPrompt(ctx, pgmodel.CreateAssistantPromptParams{
		ID:                prompt.ID,
		AppID:             prompt.AppID,
		UserID:            prompt.UserID,
		Model:             prompt.Model,
		Rounds:            int32(prompt.Rounds),
		InputTokens:       int32(prompt.Usage.InputTokens),
		CachedInputTokens: int32(prompt.Usage.CachedInputTokens),
		OutputTokens:      int32(prompt.Usage.OutputTokens),
		Prompt:            prompt.Prompt,
		Edited:            prompt.Edited,
		CreatedAt:         pgtype.Timestamp{Time: prompt.CreatedAt, Valid: true},
		UpdatedAt:         pgtype.Timestamp{Time: prompt.UpdatedAt, Valid: true},
	})
}

func (c *Client) AssistantPrompt(ctx context.Context, appID string, id string) (*model.AssistantPrompt, error) {
	row, err := c.Q.GetAssistantPrompt(ctx, pgmodel.GetAssistantPromptParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToAssistantPrompt(row), nil
}

func (c *Client) DeleteAssistantPrompt(ctx context.Context, appID string, id string) error {
	return c.Q.DeleteAssistantPrompt(ctx, pgmodel.DeleteAssistantPromptParams{
		ID:    id,
		AppID: appID,
	})
}

func (c *Client) StartAssistantPromptRound(ctx context.Context, appID string, id string, maxRounds int, updatedAt time.Time) (bool, error) {
	rows, err := c.Q.StartAssistantPromptRound(ctx, pgmodel.StartAssistantPromptRoundParams{
		ID:        id,
		AppID:     appID,
		MaxRounds: int32(maxRounds),
		UpdatedAt: pgtype.Timestamp{Time: updatedAt, Valid: true},
	})
	return rows > 0, err
}

func (c *Client) UndoAssistantPromptRound(ctx context.Context, appID string, id string) error {
	return c.Q.UndoAssistantPromptRound(ctx, pgmodel.UndoAssistantPromptRoundParams{
		ID:    id,
		AppID: appID,
	})
}

func (c *Client) AddAssistantPromptUsage(ctx context.Context, appID string, id string, usage model.AssistantUsage, edited bool, updatedAt time.Time) error {
	return c.Q.AddAssistantPromptUsage(ctx, pgmodel.AddAssistantPromptUsageParams{
		Edited:            edited,
		ID:                id,
		AppID:             appID,
		InputTokens:       int32(usage.InputTokens),
		CachedInputTokens: int32(usage.CachedInputTokens),
		OutputTokens:      int32(usage.OutputTokens),
		UpdatedAt:         pgtype.Timestamp{Time: updatedAt, Valid: true},
	})
}

func (c *Client) CountAssistantPromptsBetween(ctx context.Context, appID string, start time.Time, end time.Time) (model.AssistantPromptCount, error) {
	row, err := c.Q.CountAssistantPromptsByAppBetween(ctx, pgmodel.CountAssistantPromptsByAppBetweenParams{
		AppID:   appID,
		StartAt: pgtype.Timestamp{Time: start, Valid: true},
		EndAt:   pgtype.Timestamp{Time: end, Valid: true},
	})
	return model.AssistantPromptCount{Edited: int(row.Edited), Total: int(row.Total)}, err
}

func rowToAssistantPrompt(row pgmodel.AssistantPrompt) *model.AssistantPrompt {
	return &model.AssistantPrompt{
		ID:     row.ID,
		AppID:  row.AppID,
		UserID: row.UserID,
		Model:  row.Model,
		Prompt: row.Prompt,
		Rounds: int(row.Rounds),
		Edited: row.Edited,
		Usage: model.AssistantUsage{
			InputTokens:       int(row.InputTokens),
			CachedInputTokens: int(row.CachedInputTokens),
			OutputTokens:      int(row.OutputTokens),
		},
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
