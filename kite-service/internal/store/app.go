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

// AppGatewayRequirementsRow is the stored form of an app's gateway
// requirements. Plugin resources stay as raw "plugin_id:resource_id" pairs
// because resolving them needs the plugin registry, which the store layer has
// no access to; callers build model.AppGatewayRequirements from this.
type AppGatewayRequirementsRow struct {
	EventListenerTypes  []model.EventListenerType
	PluginResources     []string
	HasMessageInstances bool
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
	// DisabledAppIDsUpdatedSince is a cheap alternative to scanning every
	// enabled app ID: disabling an app bumps updated_at, so the gateway
	// manager can shut those connections down without a full table scan.
	DisabledAppIDsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]string, error)
	// AppGatewayRequirements reports what the app actually consumes from the
	// gateway, so it can identify with a minimal intent set.
	AppGatewayRequirements(ctx context.Context, appID string) (*AppGatewayRequirementsRow, error)
	// AppIDsWithGatewayRequirementsChangedSince reports apps whose event
	// listeners or plugin instances changed, and whose intents may therefore
	// need recomputing. Deletions are not reported; see the query for why.
	AppIDsWithGatewayRequirementsChangedSince(ctx context.Context, updatedSince time.Time) ([]string, error)

	Collaborator(ctx context.Context, appID string, userID string) (*model.AppCollaborator, error)
	CollaboratorsByApp(ctx context.Context, appID string) ([]*model.AppCollaborator, error)
	CountCollaboratorsByApp(ctx context.Context, appID string) (int, error)
	CreateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error)
	UpdateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error)
	DeleteCollaborator(ctx context.Context, appID string, userID string) error

	AppEntities(ctx context.Context, appID string) ([]*model.AppEntity, error)
}
