package store

import (
	"context"

	"github.com/kitecloud/kite/kite-common/model"
)

type UserStore interface {
	User(ctx context.Context, id string) (*model.User, error)
	UserByEmail(ctx context.Context, email string) (*model.User, error)
	UserByDiscordID(ctx context.Context, discordID string) (*model.User, error)
	UpsertUser(ctx context.Context, user *model.User) (*model.User, error)
}
