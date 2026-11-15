package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// ItemCounts counts the commands or event listeners of an app. Only the
// enabled ones count against the limit of the plan.
type ItemCounts struct {
	Total   int
	Enabled int
}

type CommandStore interface {
	CommandsByApp(ctx context.Context, appID string) ([]*model.Command, error)
	CountCommandsByApp(ctx context.Context, appID string) (ItemCounts, error)
	Command(ctx context.Context, id string) (*model.Command, error)
	CreateCommand(ctx context.Context, command *model.Command) (*model.Command, error)
	// UpdateCommand saves everything but Enabled, which only
	// UpdateCommandEnabled changes, so neither reverts the other.
	UpdateCommand(ctx context.Context, command *model.Command, change ItemChange) (*model.Command, error)
	UpdateCommandEnabled(ctx context.Context, id string, enabled bool, updatedAt time.Time) (*model.Command, error)
	EnabledCommandsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.Command, error)
	EnabledCommandIDs(ctx context.Context) ([]string, error)
	// AppsWithEnabledCommandsOver returns the number of enabled commands of
	// each app that has more than minCount.
	AppsWithEnabledCommandsOver(ctx context.Context, minCount int) (map[string]int, error)
	// DisableNewestCommands disables the enabled commands of the app but the
	// keep oldest ones, and returns how many it disabled.
	DisableNewestCommands(ctx context.Context, appID string, keep int, updatedAt time.Time) (int, error)
	// DeleteCommand keeps the command's flow in its version history.
	DeleteCommand(ctx context.Context, id string, userID string) error
}
