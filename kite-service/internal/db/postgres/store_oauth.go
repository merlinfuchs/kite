package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) CreateOAuthCode(ctx context.Context, code *model.OAuthCode) error {
	return c.Q.CreateOAuthCode(ctx, pgmodel.CreateOAuthCodeParams{
		CodeHash:      code.CodeHash,
		UserID:        code.UserID,
		ClientID:      code.ClientID,
		ClientName:    code.ClientName,
		RedirectUri:   code.RedirectURI,
		CodeChallenge: code.CodeChallenge,
		Scopes:        code.Scopes,
		AppIds:        code.AppIDs,
		Resource:      code.Resource,
		ExpiresAt:     pgtype.Timestamp{Time: code.ExpiresAt.UTC(), Valid: true},
		CreatedAt:     pgtype.Timestamp{Time: code.CreatedAt.UTC(), Valid: true},
	})
}

func (c *Client) UseOAuthCode(ctx context.Context, codeHash string, apiTokenID string, now time.Time) (*model.OAuthCode, error) {
	row, err := c.Q.UseOAuthCode(ctx, pgmodel.UseOAuthCodeParams{
		ApiTokenID: pgtype.Text{String: apiTokenID, Valid: true},
		CodeHash:   codeHash,
		Now:        pgtype.Timestamp{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToOAuthCode(row), nil
}

func (c *Client) OAuthCode(ctx context.Context, codeHash string) (*model.OAuthCode, error) {
	row, err := c.Q.GetOAuthCode(ctx, codeHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToOAuthCode(row), nil
}

func (c *Client) DeleteExpiredOAuthCodes(ctx context.Context, expiredBefore time.Time, batchSize int) (int64, error) {
	return c.Q.DeleteExpiredOAuthCodes(ctx, pgmodel.DeleteExpiredOAuthCodesParams{
		ExpiredBefore: pgtype.Timestamp{Time: expiredBefore.UTC(), Valid: true},
		BatchSize:     int32(batchSize),
	})
}

func rowToOAuthCode(row pgmodel.OauthCode) *model.OAuthCode {
	return &model.OAuthCode{
		CodeHash:      row.CodeHash,
		UserID:        row.UserID,
		ClientID:      row.ClientID,
		ClientName:    row.ClientName,
		RedirectURI:   row.RedirectUri,
		CodeChallenge: row.CodeChallenge,
		Scopes:        row.Scopes,
		AppIDs:        row.AppIds,
		Resource:      row.Resource,
		APITokenID:    null.NewString(row.ApiTokenID.String, row.ApiTokenID.Valid),
		ExpiresAt:     row.ExpiresAt.Time,
		CreatedAt:     row.CreatedAt.Time,
	}
}
