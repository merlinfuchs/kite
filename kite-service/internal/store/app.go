package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type AppUpdateOpts struct {
	ID             string
	Name           string
	Description    null.String
	DiscordToken   string
	DiscordStatus  *model.AppDiscordStatus
	Enabled        bool
	DisabledReason null.String
	UpdatedAt      time.Time
}

type AppDisableOpts struct {
	ID             string
	DisabledReason null.String
	UpdatedAt      time.Time
}

type AppStore interface {
	AppsByUser(ctx context.Context, userID string) ([]*model.App, error)
	CountAppsByUser(ctx context.Context, userID string) (int, error)
	App(ctx context.Context, id string) (*model.App, error)
	AppCredentials(ctx context.Context, id string) (*model.AppCredentials, error)
	CreateApp(ctx context.Context, app *model.App) (*model.App, error)
	UpdateApp(ctx context.Context, opts AppUpdateOpts) (*model.App, error)
	DisableApp(ctx context.Context, opts AppDisableOpts) error
	DeleteApp(ctx context.Context, id string) error
	EnabledAppIDs(ctx context.Context) ([]string, error)
	EnabledAppsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.App, error)

	Collaborator(ctx context.Context, appID string, userID string) (*model.AppCollaborator, error)
	CollaboratorsByApp(ctx context.Context, appID string) ([]*model.AppCollaborator, error)
	CountCollaboratorsByApp(ctx context.Context, appID string) (int, error)
	CreateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error)
	UpdateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error)
	DeleteCollaborator(ctx context.Context, appID string, userID string) error

	AppEntities(ctx context.Context, appID string) ([]*model.AppEntity, error)
}
