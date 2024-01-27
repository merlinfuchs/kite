package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type SessionStore interface {
	GetSession(ctx context.Context, tokenHash string) (*model.Session, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	CreateSession(ctx context.Context, session *model.Session) error
}
