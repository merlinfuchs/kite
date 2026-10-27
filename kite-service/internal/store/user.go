package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// BannedError is returned when the user is banned from the feature they use.
type BannedError struct {
	Ban *model.UserBan
}

func (e *BannedError) Error() string {
	return "user is banned from " + string(e.Ban.Feature)
}

type UserStore interface {
	User(ctx context.Context, id string) (*model.User, error)
	UserByDiscordID(ctx context.Context, discordID string) (*model.User, error)
	UserByDiscordUsername(ctx context.Context, discordUsername string) (*model.User, error)
	UpsertUser(ctx context.Context, user *model.User) (*model.User, error)
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
