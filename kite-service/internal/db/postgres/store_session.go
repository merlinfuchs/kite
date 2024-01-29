package postgres

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) GetSession(ctx context.Context, tokenHash string) (*model.Session, error) {
	row, err := c.Q.GetSession(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.Session{
		TokenHash:   row.TokenHash,
		Type:        model.SessionType(row.Type),
		UserID:      row.UserID,
		GuildIds:    row.GuildIds,
		AccessToken: row.AccessToken,
		Revoked:     row.Revoked,
		CreatedAt:   row.CreatedAt,
		ExpiresAt:   row.ExpiresAt,
	}, nil
}

func (c *Client) DeleteSession(ctx context.Context, tokenHash string) error {
	err := c.Q.DeleteSession(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}
	return nil
}

func (c *Client) CreateSession(ctx context.Context, session *model.Session) error {
	err := c.Q.CreateSession(ctx, pgmodel.CreateSessionParams{
		TokenHash:   session.TokenHash,
		Type:        string(session.Type),
		UserID:      session.UserID,
		GuildIds:    session.GuildIds,
		AccessToken: session.AccessToken,
		CreatedAt:   session.CreatedAt,
		ExpiresAt:   session.ExpiresAt,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreatePendingSession(ctx context.Context, pendingSession *model.PendingSession) error {
	err := c.Q.DeleteExpiredPendingSessions(ctx, time.Now().UTC())
	if err != nil {
		slog.With(logattr.Error(err)).Error("failed to delete expired pending sessions")
	}

	err = c.Q.CreatePendingSession(ctx, pgmodel.CreatePendingSessionParams{
		Code:      pendingSession.Code,
		CreatedAt: pendingSession.CreatedAt,
		ExpiresAt: pendingSession.ExpiresAt,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdatePendingSession(ctx context.Context, pendingSession *model.PendingSession) (*model.PendingSession, error) {
	row, err := c.Q.UpdatePendingSession(ctx, pgmodel.UpdatePendingSessionParams{
		Code:      pendingSession.Code,
		Token:     pendingSession.Token.NullString,
		ExpiresAt: pendingSession.ExpiresAt,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.PendingSession{
		Code:      row.Code,
		Token:     null.NewString(row.Token.String, row.Token.Valid),
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

func (c *Client) GetPendingSession(ctx context.Context, code string) (*model.PendingSession, error) {
	row, err := c.Q.GetPendingSession(ctx, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	if row.ExpiresAt.Before(time.Now().UTC()) {
		return nil, store.ErrNotFound
	}

	return &model.PendingSession{
		Code:      row.Code,
		Token:     null.NewString(row.Token.String, row.Token.Valid),
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}, nil
}
