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

type AppCommandsDeployResult struct {
	ID string
	// DeployAt is the app's CommandsDeployAt when the deploy started.
	DeployAt time.Time
	// DeployedHash is the hash of the commands in Discord, or null if unknown.
	DeployedHash null.String
	Error        *model.CommandsDeployError
}

// AppGatewayRequirementsRow is the stored form of an app's gateway
// requirements. Plugin resources stay as raw "plugin_id:resource_id" pairs
// because resolving them needs the plugin registry, which the store layer has
// no access to; callers build model.AppGatewayRequirements from this.
type AppGatewayRequirementsRow struct {
	EventListenerTypes  []model.EventListenerType
	PluginResources     []string
	HasMessageInstances bool
	UsesVoice           bool
}

type AppStore interface {
	AppsByUser(ctx context.Context, userID string) ([]*model.App, error)
	CountAppsByUser(ctx context.Context, userID string) (int, error)
	App(ctx context.Context, id string) (*model.App, error)
	AppCredentials(ctx context.Context, id string) (*model.AppCredentials, error)
	// AppDiscordUserInstallCount is null until the gateway first fetched it.
	AppDiscordUserInstallCount(ctx context.Context, id string) (null.Int, error)
	CreateApp(ctx context.Context, app *model.App) (*model.App, error)
	UpdateApp(ctx context.Context, opts AppUpdateOpts) (*model.App, error)
	// DisableApp reports whether the app was enabled before.
	DisableApp(ctx context.Context, opts AppDisableOpts) (bool, error)
	// PauseAppForCredits and UnpauseAppForCredits report whether the app
	// wasn't paused or was paused before.
	PauseAppForCredits(ctx context.Context, appID string, now time.Time) (bool, error)
	UnpauseAppForCredits(ctx context.Context, appID string, now time.Time) (bool, error)
	CreditsPausedAppIDs(ctx context.Context) ([]string, error)
	DeleteApp(ctx context.Context, id string) error
	EnabledAppIDs(ctx context.Context) ([]string, error)
	EnabledAppsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.App, error)
	// DisabledAppIDsUpdatedSince is a cheap alternative to scanning every
	// enabled app ID: disabling an app bumps updated_at, so the gateway
	// manager can shut those connections down without a full table scan.
	DisabledAppIDsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]string, error)
	UpdateAppCreditNotices(ctx context.Context, appID string, enabled bool, updatedAt time.Time) (*model.App, error)
	// AppGatewayRequirements reports what the app actually consumes from the
	// gateway, so it can identify with a minimal intent set.
	AppGatewayRequirements(ctx context.Context, appID string) (*AppGatewayRequirementsRow, error)
	// AppIDsWithGatewayRequirementsChangedSince reports apps whose event
	// listeners or plugin instances changed, and whose intents may therefore
	// need recomputing. Deletions are not reported; see the query for why.
	AppIDsWithGatewayRequirementsChangedSince(ctx context.Context, updatedSince time.Time) ([]string, error)
	// QueueAppCommandsDeploy deploys the app's commands again in a few
	// seconds. Changes to commands and plugins do so by themselves. Forcing
	// sends them even if they look unchanged.
	QueueAppCommandsDeploy(ctx context.Context, appID string, force bool) error
	AppIDsWithDueCommandsDeploy(ctx context.Context, limit int) ([]string, error)
	FinishAppCommandsDeploy(ctx context.Context, res AppCommandsDeployResult) error
	// DelayAppCommandsDeploy tries again in a few minutes.
	DelayAppCommandsDeploy(ctx context.Context, appID string, err *model.CommandsDeployError) error
	UpdateAppDiscordCounts(ctx context.Context, appID string, guildCount int, userInstallCount int) error
	UpdateAppDiscordVerifyKey(ctx context.Context, appID string, verifyKey string) error

	Collaborator(ctx context.Context, appID string, userID string) (*model.AppCollaborator, error)
	CollaboratorsByApp(ctx context.Context, appID string) ([]*model.AppCollaborator, error)
	CreateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error)
	UpdateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error)
	DeleteCollaborator(ctx context.Context, appID string, userID string) error

	CollaboratorInvitesByApp(ctx context.Context, appID string, now time.Time) ([]*model.AppCollaboratorInvite, error)
	// CountCollaboratorsAndInvitesByApp counts collaborators and pending invites,
	// which both take up a seat.
	CountCollaboratorsAndInvitesByApp(ctx context.Context, appID string, now time.Time) (int, error)
	UpsertCollaboratorInvite(ctx context.Context, invite *model.AppCollaboratorInvite) (*model.AppCollaboratorInvite, error)
	DeleteCollaboratorInvite(ctx context.Context, appID string, discordUsername string) error
	DeleteExpiredCollaboratorInvites(ctx context.Context, now time.Time) error
	CollaboratorInvite(ctx context.Context, appID string, discordUsername string) (*model.AppCollaboratorInvite, error)
	// CollaboratorInvitesByUser returns the pending invites of the user, by
	// their ID or, for invites from before they used Kite, their username.
	CollaboratorInvitesByUser(ctx context.Context, userID string, discordUsername string, now time.Time) ([]*model.AppCollaboratorInvite, error)
	// AcceptCollaboratorInvite deletes the user's invite to the app and adds
	// them as a collaborator. It returns store.ErrNotFound without an invite.
	AcceptCollaboratorInvite(ctx context.Context, appID string, userID string, discordUsername string, now time.Time) (*model.AppCollaborator, error)
	// DeclineCollaboratorInvite deletes the user's invite to the app. It
	// returns store.ErrNotFound without an invite.
	DeclineCollaboratorInvite(ctx context.Context, appID string, userID string, discordUsername string) error

	CreateAppAuditLogEntry(ctx context.Context, entry *model.AppAuditLogEntry) error
	// AppAuditLogEntries returns the newest entries of the app that match
	// the filter.
	AppAuditLogEntries(ctx context.Context, appID string, filter AppAuditLogFilter) ([]*model.AppAuditLogEntry, error)
	DeleteAppAuditLogEntriesBefore(ctx context.Context, before time.Time) error

	AppEntities(ctx context.Context, appID string) ([]*model.AppEntity, error)
	// HasAppFlows reports whether the app has an enabled command or event
	// listener, or a posted message with flows.
	HasAppFlows(ctx context.Context, appID string) (bool, error)
}

type AppAuditLogFilter struct {
	// BeforeID only returns entries before the one with the ID, unless it's 0.
	BeforeID int64
	// UserID only returns entries of the user, unless it's empty.
	UserID string
	// Kinds only returns entries whose action starts with one of them, like
	// "command" for "command.update", unless it's empty.
	Kinds []string
	Limit int
}
