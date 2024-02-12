package postgres

import (
	"context"
	"database/sql"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func (c *Client) UpsertUser(ctx context.Context, user *model.User) error {
	_, err := c.Q.UpsertUser(ctx, pgmodel.UpsertUserParams{
		ID:            string(user.ID),
		Username:      user.Username,
		Discriminator: sql.NullString{String: user.Discriminator, Valid: user.Discriminator != ""},
		Avatar:        sql.NullString{String: user.Avatar, Valid: user.Avatar != ""},
		GlobalName:    sql.NullString{String: user.GlobalName, Valid: user.GlobalName != ""},
		PublicFlags:   int32(user.PublicFlags),
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	})
	return err
}

func (c *Client) GetUser(ctx context.Context, userID distype.Snowflake) (*model.User, error) {
	row, err := c.Q.GetUser(ctx, string(userID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.User{
		ID:            distype.Snowflake(row.ID),
		Username:      row.Username,
		Discriminator: row.Discriminator.String,
		Avatar:        row.Avatar.String,
		GlobalName:    row.GlobalName.String,
		PublicFlags:   int(row.PublicFlags),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}
