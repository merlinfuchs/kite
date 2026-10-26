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
	UpdateCommand(ctx context.Context, command *model.Command, change ItemChange) (*model.Command, error)
	EnabledCommandsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.Command, error)
	EnabledCommandIDs(ctx context.Context) ([]string, error)
	// DeleteCommand keeps the command's flow in its version history.
	DeleteCommand(ctx context.Context, id string, userID string) error
}
