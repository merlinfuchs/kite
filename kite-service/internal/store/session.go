package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type SessionStore interface {
	CreateSession(ctx context.Context, session *model.Session) (*model.Session, error)
	DeleteSession(ctx context.Context, keyHash string) error
	Session(ctx context.Context, keyHash string) (*model.Session, error)
	UpdateSessionExpiry(ctx context.Context, keyHash string, expiresAt time.Time) error
}
