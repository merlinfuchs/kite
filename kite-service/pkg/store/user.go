package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type UserStore interface {
	UpsertUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, userID string) (*model.User, error)
}
