package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) User(ctx context.Context, id string) (*model.User, error) {
	row, err := c.Q.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToUser(row), nil
}

func (c *Client) UserByDiscordID(ctx context.Context, discordID string) (*model.User, error) {
	row, err := c.Q.GetUserByDiscordID(ctx, discordID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToUser(row), nil
}

func (c *Client) UpsertUser(ctx context.Context, user *model.User) (*model.User, error) {
	row, err := c.Q.UpsertUser(ctx, pgmodel.UpsertUserParams{
		ID:              user.ID,
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		DiscordID:       user.DiscordID,
		DiscordUsername: user.DiscordUsername,
		DiscordAvatar: pgtype.Text{
			String: user.DiscordAvatar.String,
			Valid:  user.DiscordAvatar.Valid,
		},
		SignupReferrer:    textFromNullString(user.SignupSource.Referrer),
		SignupUtmSource:   textFromNullString(user.SignupSource.UTMSource),
		SignupUtmMedium:   textFromNullString(user.SignupSource.UTMMedium),
		SignupUtmCampaign: textFromNullString(user.SignupSource.UTMCampaign),
		SignupLandingPage: textFromNullString(user.SignupSource.LandingPage),
		CreatedAt: pgtype.Timestamp{
			Time:  user.CreatedAt.UTC(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:  user.UpdatedAt.UTC(),
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return rowToUser(row), nil
}

func rowToUser(row pgmodel.User) *model.User {
	return &model.User{
		ID:              row.ID,
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		DiscordID:       row.DiscordID,
		DiscordUsername: row.DiscordUsername,
		DiscordAvatar:   null.NewString(row.DiscordAvatar.String, row.DiscordAvatar.Valid),
		SignupSource: model.SignupSource{
			Referrer:    null.NewString(row.SignupReferrer.String, row.SignupReferrer.Valid),
			UTMSource:   null.NewString(row.SignupUtmSource.String, row.SignupUtmSource.Valid),
			UTMMedium:   null.NewString(row.SignupUtmMedium.String, row.SignupUtmMedium.Valid),
			UTMCampaign: null.NewString(row.SignupUtmCampaign.String, row.SignupUtmCampaign.Valid),
			LandingPage: null.NewString(row.SignupLandingPage.String, row.SignupLandingPage.Valid),
		},
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func textFromNullString(s null.String) pgtype.Text {
	return pgtype.Text{String: s.String, Valid: s.Valid}
}
