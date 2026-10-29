package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type UserStore interface {
	User(ctx context.Context, id string) (*model.User, error)
	UserByDiscordID(ctx context.Context, discordID string) (*model.User, error)
	UserByDiscordUsername(ctx context.Context, discordUsername string) (*model.User, error)
	UpsertUser(ctx context.Context, user *model.User) (*model.User, error)
	// UserBan returns the user's ban from the feature, or ErrNotFound if
	// they aren't banned from it.
	UserBan(ctx context.Context, userID string, feature model.UserBanFeature) (*model.UserBan, error)
	// UserBans returns all bans, newest first, with their user.
	UserBans(ctx context.Context) ([]*model.UserBan, error)
	// BanUser bans the user from the feature, replacing their ban from it. A
	// ban from the marketplace rejects the pending reviews of their bundles
	// with its reason.
	BanUser(ctx context.Context, ban *model.UserBan) error
	// UnbanUser lifts the user's ban from the feature. It returns ErrNotFound
	// if they aren't banned from it.
	UnbanUser(ctx context.Context, userID string, feature model.UserBanFeature) error
}
