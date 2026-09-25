package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type FlowAIPromptStore interface {
	CreateFlowAIPrompt(ctx context.Context, prompt *model.FlowAIPrompt) error
	FlowAIPrompt(ctx context.Context, appID string, id string) (*model.FlowAIPrompt, error)
	// AddFlowAIPromptRound records another model call made for the prompt.
	AddFlowAIPromptRound(ctx context.Context, appID string, id string, usage model.FlowAIUsage, updatedAt time.Time) error
	CountFlowAIPromptsBetween(ctx context.Context, appID string, start time.Time, end time.Time) (int, error)
}
