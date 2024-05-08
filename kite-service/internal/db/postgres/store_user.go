package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.UserStore = (*Client)(nil)

func (c *Client) UpsertUser(ctx context.Context, user *model.User) error {
	_, err := c.Q.UpsertUser(ctx, pgmodel.UpsertUserParams{
		ID:            string(user.ID),
		Username:      user.Username,
		Email:         user.Email,
		Discriminator: nullStringToText(user.Discriminator),
		Avatar:        nullStringToText(user.Avatar),
		GlobalName:    nullStringToText(user.GlobalName),
		PublicFlags:   int32(user.PublicFlags),
		CreatedAt:     timeToTimestamp(user.CreatedAt),
		UpdatedAt:     timeToTimestamp(user.UpdatedAt),
	})
	return err
}

func (c *Client) GetUser(ctx context.Context, userID distype.Snowflake) (*model.User, error) {
	row, err := c.Q.GetUser(ctx, string(userID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.User{
		ID:            distype.Snowflake(row.ID),
		Username:      row.Username,
		Email:         row.Email,
		Discriminator: textToNullString(row.Discriminator),
		Avatar:        textToNullString(row.Avatar),
		GlobalName:    textToNullString(row.GlobalName),
		PublicFlags:   int(row.PublicFlags),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}
