package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type CommandStore interface {
	CommandsByApp(ctx context.Context, appID string) ([]*model.Command, error)
	CountCommandsByApp(ctx context.Context, appID string) (int, error)
	Command(ctx context.Context, id string) (*model.Command, error)
	CreateCommand(ctx context.Context, command *model.Command) (*model.Command, error)
	UpdateCommand(ctx context.Context, command *model.Command) (*model.Command, error)
	UpdateCommandsLastDeployedAt(ctx context.Context, appID string, lastDeployedAt time.Time) error
	EnabledCommandsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.Command, error)
	EnabledCommandIDs(ctx context.Context) ([]string, error)
	DeleteCommand(ctx context.Context, id string) error
	DinstinctAppIDsWithUndeployedCommands(ctx context.Context) ([]string, error)
	// MoveCommand swaps the position of the command with its neighbor in the
	// given direction ("up" or "down") and returns the app's commands in
	// their new order. It is a no-op if the command is already at that edge.
	MoveCommand(ctx context.Context, appID string, id string, direction string) ([]*model.Command, error)
}
