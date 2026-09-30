package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type AssistantPromptStore interface {
	CreateAssistantPrompt(ctx context.Context, prompt *model.AssistantPrompt) error
	AssistantPrompt(ctx context.Context, appID string, id string) (*model.AssistantPrompt, error)
	DeleteAssistantPrompt(ctx context.Context, appID string, id string) error
	// StartAssistantPromptRound records another model call for the prompt, unless
	// it already had maxRounds. It returns whether it did.
	StartAssistantPromptRound(ctx context.Context, appID string, id string, maxRounds int, updatedAt time.Time) (bool, error)
	// UndoAssistantPromptRound takes back a round the model didn't answer.
	UndoAssistantPromptRound(ctx context.Context, appID string, id string) error
	// AddAssistantPromptUsage records the usage of a model call for the prompt,
	// and marks it as unedited if edited is false.
	AddAssistantPromptUsage(ctx context.Context, appID string, id string, usage model.AssistantUsage, edited bool, updatedAt time.Time) error
	CountAssistantPromptsBetween(ctx context.Context, appID string, start time.Time, end time.Time) (model.AssistantPromptCount, error)
}
