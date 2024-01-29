package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type SessionStore interface {
	GetSession(ctx context.Context, tokenHash string) (*model.Session, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	CreateSession(ctx context.Context, session *model.Session) error
	CreatePendingSession(ctx context.Context, pendingSession *model.PendingSession) error
	UpdatePendingSession(ctx context.Context, pendingSession *model.PendingSession) (*model.PendingSession, error)
	GetPendingSession(ctx context.Context, code string) (*model.PendingSession, error)
}
