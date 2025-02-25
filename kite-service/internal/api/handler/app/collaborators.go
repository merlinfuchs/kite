package app

import (
	"errors"
	"fmt"
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

func (h *AppHandler) HandleAppCollaboratorCreate(c *handler.Context, req wire.AppCollaboratorCreateRequest) (*wire.AppCollaboratorCreateResponse, error) {
	if !c.UserAppRole.CanManageCollaborators() {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to add collaborators to this app")
	}

	features := h.planManager.AppFeatures(c.Context(), c.App.ID)
	if features.MaxCollaborators != 0 {
		collaboratorCount, err := h.appStore.CountCollaboratorsByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count collaborators: %w", err)
		}

		// We count the owner as a collaborator
		if (collaboratorCount + 1) >= features.MaxCollaborators {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of collaborators (%d) reached", features.MaxCollaborators))
		}
	}

	user, err := h.userStore.UserByDiscordID(c.Context(), req.DiscordUserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_user", "User not found")
		}
		return nil, err
	}

	if user.ID == c.App.OwnerUserID {
		return nil, handler.ErrBadRequest("cannot_add_owner", "Cannot add owner as collaborator")
	}

	collaborator, err := h.appStore.CreateCollaborator(c.Context(), &model.AppCollaborator{
		AppID:     c.App.ID,
		UserID:    user.ID,
		Role:      model.AppCollaboratorRole(req.Role),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	return wire.CollaboratorToWire(collaborator), nil
}

func (h *AppHandler) HandleAppCollaboratorDelete(c *handler.Context) (*wire.AppCollaboratorDeleteResponse, error) {
	if !c.UserAppRole.CanManageCollaborators() {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to delete collaborators from this app")
	}

	userID := c.Param("userID")

	err := h.appStore.DeleteCollaborator(c.Context(), c.App.ID, userID)
	if err != nil {
		return nil, err
	}

	return &wire.AppCollaboratorDeleteResponse{}, nil
}
