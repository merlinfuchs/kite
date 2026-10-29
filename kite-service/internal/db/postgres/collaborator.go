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
)

func (c *Client) Collaborator(ctx context.Context, appID string, userID string) (*model.AppCollaborator, error) {
	row, err := c.Q.GetCollaborator(ctx, pgmodel.GetCollaboratorParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToCollaborator(row.Collaborator, &row.User), nil
}

func (c *Client) CollaboratorsByApp(ctx context.Context, appID string) ([]*model.AppCollaborator, error) {
	rows, err := c.Q.GetCollaboratorsByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	collaborators := make([]*model.AppCollaborator, len(rows))
	for i, row := range rows {
		collaborators[i] = rowToCollaborator(row.Collaborator, &row.User)
	}

	return collaborators, nil
}

func (c *Client) CreateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error) {
	row, err := c.Q.CreateCollaborator(ctx, pgmodel.CreateCollaboratorParams{
		AppID:     collaborator.AppID,
		UserID:    collaborator.UserID,
		Role:      string(collaborator.Role),
		CreatedAt: pgtype.Timestamp{Time: collaborator.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: collaborator.UpdatedAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToCollaborator(row, nil), nil
}

func (c *Client) UpdateCollaborator(ctx context.Context, collaborator *model.AppCollaborator) (*model.AppCollaborator, error) {
	row, err := c.Q.UpdateCollaborator(ctx, pgmodel.UpdateCollaboratorParams{
		AppID:     collaborator.AppID,
		UserID:    collaborator.UserID,
		Role:      string(collaborator.Role),
		UpdatedAt: pgtype.Timestamp{Time: collaborator.UpdatedAt, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToCollaborator(row, nil), nil
}

func (c *Client) DeleteCollaborator(ctx context.Context, appID string, userID string) error {
	err := c.Q.DeleteCollaborator(ctx, pgmodel.DeleteCollaboratorParams{
		AppID:  appID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CollaboratorInvitesByApp(ctx context.Context, appID string, now time.Time) ([]*model.AppCollaboratorInvite, error) {
	rows, err := c.Q.GetCollaboratorInvitesByApp(ctx, pgmodel.GetCollaboratorInvitesByAppParams{
		AppID: appID,
		Now:   pgtype.Timestamp{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	invites := make([]*model.AppCollaboratorInvite, len(rows))
	for i, row := range rows {
		invites[i] = rowToCollaboratorInvite(row)
	}

	return invites, nil
}

func (c *Client) CountCollaboratorsAndInvitesByApp(ctx context.Context, appID string, now time.Time) (int, error) {
	count, err := c.Q.CountCollaboratorsAndInvitesByApp(ctx, pgmodel.CountCollaboratorsAndInvitesByAppParams{
		AppID: appID,
		Now:   pgtype.Timestamp{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (c *Client) UpsertCollaboratorInvite(ctx context.Context, invite *model.AppCollaboratorInvite) (*model.AppCollaboratorInvite, error) {
	row, err := c.Q.UpsertCollaboratorInvite(ctx, pgmodel.UpsertCollaboratorInviteParams{
		AppID:           invite.AppID,
		DiscordUsername: invite.DiscordUsername,
		Role:            string(invite.Role),
		CreatedAt:       pgtype.Timestamp{Time: invite.CreatedAt.UTC(), Valid: true},
		ExpiresAt:       pgtype.Timestamp{Time: invite.ExpiresAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToCollaboratorInvite(row), nil
}

func (c *Client) DeleteCollaboratorInvite(ctx context.Context, appID string, discordUsername string) error {
	return c.Q.DeleteCollaboratorInvite(ctx, pgmodel.DeleteCollaboratorInviteParams{
		AppID:           appID,
		DiscordUsername: discordUsername,
	})
}

func (c *Client) ClaimCollaboratorInvites(ctx context.Context, discordUsername string, userID string, now time.Time) error {
	return c.Q.ClaimCollaboratorInvites(ctx, pgmodel.ClaimCollaboratorInvitesParams{
		DiscordUsername: discordUsername,
		UserID:          userID,
		Now:             pgtype.Timestamp{Time: now.UTC(), Valid: true},
	})
}

func rowToCollaboratorInvite(row pgmodel.CollaboratorInvite) *model.AppCollaboratorInvite {
	return &model.AppCollaboratorInvite{
		AppID:           row.AppID,
		DiscordUsername: row.DiscordUsername,
		Role:            model.AppCollaboratorRole(row.Role),
		CreatedAt:       row.CreatedAt.Time,
		ExpiresAt:       row.ExpiresAt.Time,
	}
}

func rowToCollaborator(row pgmodel.Collaborator, user *pgmodel.User) *model.AppCollaborator {
	var userModel *model.User
	if user != nil {
		userModel = rowToUser(*user)
	}

	return &model.AppCollaborator{
		AppID:     row.AppID,
		UserID:    row.UserID,
		User:      userModel,
		Role:      model.AppCollaboratorRole(row.Role),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
