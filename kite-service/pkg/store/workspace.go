package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type WorkspaceStore interface {
	GetWorkspace(ctx context.Context, id string, guildID string) (*model.Workspace, error)
	GetWorkspacesForGuild(ctx context.Context, guildID string) ([]model.Workspace, error)
	CreateWorkspace(ctx context.Context, workspace model.Workspace) (*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace model.Workspace) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, id string, guildID string) error
}
