package postgres

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) GetSession(ctx context.Context, tokenHash string) (*model.Session, error) {
	row, err := c.Q.GetSession(ctx, tokenHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	guildIDs := make([]distype.Snowflake, len(row.GuildIds))
	for i, id := range row.GuildIds {
		guildIDs[i] = distype.Snowflake(id)
	}

	return &model.Session{
		TokenHash:   row.TokenHash,
		Type:        model.SessionType(row.Type),
		UserID:      distype.Snowflake(row.UserID),
		GuildIds:    guildIDs,
		AccessToken: row.AccessToken,
		Revoked:     row.Revoked,
		CreatedAt:   row.CreatedAt.Time,
		ExpiresAt:   row.ExpiresAt.Time,
	}, nil
}

func (c *Client) DeleteSession(ctx context.Context, tokenHash string) error {
	err := c.Q.DeleteSession(ctx, tokenHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}
	return nil
}

func (c *Client) CreateSession(ctx context.Context, session *model.Session) error {
	guildIDs := make([]string, len(session.GuildIds))
	for i, id := range session.GuildIds {
		guildIDs[i] = string(id)
	}

	err := c.Q.CreateSession(ctx, pgmodel.CreateSessionParams{
		TokenHash:   session.TokenHash,
		Type:        string(session.Type),
		UserID:      string(session.UserID),
		GuildIds:    guildIDs,
		AccessToken: session.AccessToken,
		CreatedAt:   timeToTimestamp(session.CreatedAt),
		ExpiresAt:   timeToTimestamp(session.ExpiresAt),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreatePendingSession(ctx context.Context, pendingSession *model.PendingSession) error {
	err := c.Q.DeleteExpiredPendingSessions(ctx, timeToTimestamp(time.Now().UTC()))
	if err != nil {
		slog.With(logattr.Error(err)).Error("failed to delete expired pending sessions")
	}

	err = c.Q.CreatePendingSession(ctx, pgmodel.CreatePendingSessionParams{
		Code:      pendingSession.Code,
		CreatedAt: timeToTimestamp(pendingSession.CreatedAt),
		ExpiresAt: timeToTimestamp(pendingSession.ExpiresAt),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdatePendingSession(ctx context.Context, pendingSession *model.PendingSession) (*model.PendingSession, error) {
	row, err := c.Q.UpdatePendingSession(ctx, pgmodel.UpdatePendingSessionParams{
		Code:      pendingSession.Code,
		Token:     nullStringToText(pendingSession.Token),
		ExpiresAt: timeToTimestamp(pendingSession.ExpiresAt),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.PendingSession{
		Code:      row.Code,
		Token:     null.NewString(row.Token.String, row.Token.Valid),
		CreatedAt: row.CreatedAt.Time,
		ExpiresAt: row.ExpiresAt.Time,
	}, nil
}

func (c *Client) GetPendingSession(ctx context.Context, code string) (*model.PendingSession, error) {
	row, err := c.Q.GetPendingSession(ctx, code)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	if row.ExpiresAt.Time.Before(time.Now().UTC()) {
		return nil, store.ErrNotFound
	}

	return &model.PendingSession{
		Code:      row.Code,
		Token:     null.NewString(row.Token.String, row.Token.Valid),
		CreatedAt: row.CreatedAt.Time,
		ExpiresAt: row.ExpiresAt.Time,
	}, nil
}
