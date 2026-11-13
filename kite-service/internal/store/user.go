package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type UserStore interface {
	User(ctx context.Context, id string) (*model.User, error)
	UserByDiscordID(ctx context.Context, discordID string) (*model.User, error)
	// UserByAppDiscordID returns the owner of the app with the Discord ID.
	UserByAppDiscordID(ctx context.Context, discordID string) (*model.User, error)
	UserByDiscordUsername(ctx context.Context, discordUsername string) (*model.User, error)
	UpsertUser(ctx context.Context, user *model.User) (*model.User, error)
	// UserBan returns the user's ban from the feature, or ErrNotFound if
	// they aren't banned from it.
	UserBan(ctx context.Context, userID string, feature model.UserBanFeature) (*model.UserBan, error)
	// UserBans returns all bans, newest first, with their user.
	UserBans(ctx context.Context) ([]*model.UserBan, error)
	// BanUser bans the user from the feature, replacing their ban from it. A
	// ban from the marketplace or Kite rejects the pending reviews of their
	// bundles with its reason, and a ban from Kite stops their apps.
	BanUser(ctx context.Context, ban *model.UserBan) error
	// UnbanUser lifts the user's ban from the feature. It returns ErrNotFound
	// if they aren't banned from it. Lifting a ban from Kite starts the apps
	// it stopped again.
	UnbanUser(ctx context.Context, userID string, feature model.UserBanFeature) error
	// UserBadgeGrants returns all badges admins gave, newest first, with their
	// user.
	UserBadgeGrants(ctx context.Context) ([]*model.UserBadgeGrant, error)
	// UserBadgeGrantsByUserIDs returns the badges admins gave the users.
	UserBadgeGrantsByUserIDs(ctx context.Context, userIDs []string) ([]*model.UserBadgeGrant, error)
	// GrantUserBadge gives the user the badge, which does nothing if they have
	// it already.
	GrantUserBadge(ctx context.Context, grant *model.UserBadgeGrant) error
	// RevokeUserBadge takes the badge away again. It returns ErrNotFound if
	// the user doesn't have it.
	RevokeUserBadge(ctx context.Context, userID string, kind model.UserBadgeKind) error
}
