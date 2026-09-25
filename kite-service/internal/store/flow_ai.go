package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type FlowAIPromptStore interface {
	CreateFlowAIPrompt(ctx context.Context, prompt *model.FlowAIPrompt) error
	FlowAIPrompt(ctx context.Context, appID string, id string) (*model.FlowAIPrompt, error)
	MarkFlowAIPromptUnedited(ctx context.Context, appID string, id string) error
	DeleteFlowAIPrompt(ctx context.Context, appID string, id string) error
	// StartFlowAIPromptRound records another model call for the prompt, unless
	// it already had maxRounds. It returns whether it did.
	StartFlowAIPromptRound(ctx context.Context, appID string, id string, maxRounds int, updatedAt time.Time) (bool, error)
	AddFlowAIPromptUsage(ctx context.Context, appID string, id string, usage model.FlowAIUsage, updatedAt time.Time) error
	CountFlowAIPromptsBetween(ctx context.Context, appID string, start time.Time, end time.Time) (model.FlowAIPromptCount, error)
}
