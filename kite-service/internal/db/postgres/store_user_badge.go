package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) UserBadgeGrants(ctx context.Context) ([]*model.UserBadgeGrant, error) {
	rows, err := c.Q.GetUserBadgeGrants(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*model.UserBadgeGrant, len(rows))
	for i, row := range rows {
		res[i] = rowToUserBadgeGrant(row.UserBadgeGrant)
		res[i].User = rowToUser(row.User)
	}
	return res, nil
}

func (c *Client) UserBadgeGrantsByUserIDs(ctx context.Context, userIDs []string) ([]*model.UserBadgeGrant, error) {
	rows, err := c.Q.GetUserBadgeGrantsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	res := make([]*model.UserBadgeGrant, len(rows))
	for i, row := range rows {
		res[i] = rowToUserBadgeGrant(row)
	}
	return res, nil
}

func (c *Client) GrantUserBadge(ctx context.Context, grant *model.UserBadgeGrant) error {
	return c.Q.UpsertUserBadgeGrant(ctx, pgmodel.UpsertUserBadgeGrantParams{
		UserID:          grant.UserID,
		Badge:           string(grant.Kind),
		GrantedByUserID: pgtype.Text{String: grant.GrantedByUserID.String, Valid: grant.GrantedByUserID.Valid},
		CreatedAt:       pgtype.Timestamp{Time: grant.CreatedAt.UTC(), Valid: true},
	})
}

func (c *Client) RevokeUserBadge(ctx context.Context, userID string, kind model.UserBadgeKind) error {
	rows, err := c.Q.DeleteUserBadgeGrant(ctx, pgmodel.DeleteUserBadgeGrantParams{
		UserID: userID,
		Badge:  string(kind),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func rowToUserBadgeGrant(row pgmodel.UserBadgeGrant) *model.UserBadgeGrant {
	return &model.UserBadgeGrant{
		UserID:          row.UserID,
		Kind:            model.UserBadgeKind(row.Badge),
		GrantedByUserID: null.NewString(row.GrantedByUserID.String, row.GrantedByUserID.Valid),
		CreatedAt:       row.CreatedAt.Time,
	}
}
