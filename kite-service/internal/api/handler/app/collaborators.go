package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (h *AppHandler) HandleAppCollaboratorsList(c *handler.Context) (*wire.AppCollaboratorListResponse, error) {
	collaborators, err := h.appStore.CollaboratorsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, err
	}

	ownerUser, err := h.userStore.User(c.Context(), c.App.OwnerUserID)
	if err != nil {
		return nil, err
	}

	res := make(wire.AppCollaboratorListResponse, len(collaborators)+1)
	res[0] = &wire.AppCollaborator{
		User:      *wire.UserToWire(ownerUser, true),
		Role:      string(model.AppCollaboratorRoleOwner),
		CreatedAt: c.App.CreatedAt,
		UpdatedAt: c.App.CreatedAt,
	}

	for i, collaborator := range collaborators {
		res[i+1] = wire.CollaboratorToWire(collaborator)
	}

	return &res, nil
}

// collaboratorInviteTTL is how long a user has to log in to Kite after being
// added by a Discord username that no Kite user has.
const collaboratorInviteTTL = 7 * 24 * time.Hour

func (h *AppHandler) HandleAppCollaboratorCreate(c *handler.Context, req wire.AppCollaboratorCreateRequest) (*wire.AppCollaboratorCreateResponse, error) {
	if !c.UserAppRole.CanManageCollaborators() {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to add collaborators to this app")
	}

	now := time.Now().UTC()

	if c.Features.MaxCollaborators != 0 {
		count, err := h.appStore.CountCollaboratorsAndInvitesByApp(c.Context(), c.App.ID, now)
		if err != nil {
			return nil, fmt.Errorf("failed to count collaborators: %w", err)
		}

		// We count the owner as a collaborator
		if (count + 1) >= c.Features.MaxCollaborators {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of collaborators (%d) reached", c.Features.MaxCollaborators))
		}
	}

	isUserID := wire.DiscordUserIDRegex.MatchString(req.DiscordUser)
	discordUsername := strings.ToLower(strings.TrimPrefix(req.DiscordUser, "@"))

	var user *model.User
	var err error
	if isUserID {
		user, err = h.userStore.UserByDiscordID(c.Context(), req.DiscordUser)
	} else {
		user, err = h.userStore.UserByDiscordUsername(c.Context(), discordUsername)
	}
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			return nil, err
		}
		if isUserID {
			return nil, handler.ErrNotFound("unknown_user", "User not found. They need to log in to Kite once, or add them by their Discord username instead.")
		}

		invite, err := h.appStore.UpsertCollaboratorInvite(c.Context(), &model.AppCollaboratorInvite{
			AppID:           c.App.ID,
			DiscordUsername: discordUsername,
			Role:            model.AppCollaboratorRole(req.Role),
			CreatedAt:       now,
			ExpiresAt:       now.Add(collaboratorInviteTTL),
		})
		if err != nil {
			return nil, err
		}

		return &wire.AppCollaboratorCreateResponse{
			Invite: wire.CollaboratorInviteToWire(invite),
		}, nil
	}

	if user.ID == c.App.OwnerUserID {
		return nil, handler.ErrBadRequest("cannot_add_owner", "Cannot add owner as collaborator")
	}

	_, err = h.appStore.Collaborator(c.Context(), c.App.ID, user.ID)
	if err == nil {
		return nil, handler.ErrBadRequest("already_collaborator", "This user is already a collaborator")
	} else if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	collaborator, err := h.appStore.CreateCollaborator(c.Context(), &model.AppCollaborator{
		AppID:     c.App.ID,
		UserID:    user.ID,
		Role:      model.AppCollaboratorRole(req.Role),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, err
	}
	collaborator.User = user

	return &wire.AppCollaboratorCreateResponse{
		Collaborator: wire.CollaboratorToWire(collaborator),
	}, nil
}

func (h *AppHandler) HandleAppCollaboratorDelete(c *handler.Context) (*wire.AppCollaboratorDeleteResponse, error) {
	userID := c.Param("userID")

	// Collaborators can always remove themselves to leave the app
	if !c.UserAppRole.CanManageCollaborators() && userID != c.Session.UserID {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to delete collaborators from this app")
	}

	err := h.appStore.DeleteCollaborator(c.Context(), c.App.ID, userID)
	if err != nil {
		return nil, err
	}

	return &wire.AppCollaboratorDeleteResponse{}, nil
}

func (h *AppHandler) HandleAppCollaboratorInviteList(c *handler.Context) (*wire.AppCollaboratorInviteListResponse, error) {
	invites, err := h.appStore.CollaboratorInvitesByApp(c.Context(), c.App.ID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	res := make(wire.AppCollaboratorInviteListResponse, len(invites))
	for i, invite := range invites {
		res[i] = wire.CollaboratorInviteToWire(invite)
	}

	return &res, nil
}

func (h *AppHandler) HandleAppCollaboratorInviteDelete(c *handler.Context) (*wire.AppCollaboratorInviteDeleteResponse, error) {
	if !c.UserAppRole.CanManageCollaborators() {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to delete collaborators from this app")
	}

	err := h.appStore.DeleteCollaboratorInvite(c.Context(), c.App.ID, c.Param("discordUsername"))
	if err != nil {
		return nil, err
	}

	return &wire.AppCollaboratorInviteDeleteResponse{}, nil
}
