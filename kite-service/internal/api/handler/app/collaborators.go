package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
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

// collaboratorInviteTTL is how long a user has to accept an invite.
const collaboratorInviteTTL = 7 * 24 * time.Hour

// HandleAppCollaboratorCreate invites a user to the app. They become a
// collaborator once they accept the invite, which they see in their account
// menu.
func (h *AppHandler) HandleAppCollaboratorCreate(c *handler.Context, req wire.AppCollaboratorCreateRequest) (*wire.AppCollaboratorCreateResponse, error) {
	if !c.UserAppRole.CanManageCollaborators() {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to invite collaborators to this app")
	}

	role := model.AppCollaboratorRole(req.Role)
	if !c.UserAppRole.CanInviteRole(role) {
		return nil, handler.ErrForbidden("missing_permissions", "Only the owner of the app can invite admins")
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
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}
	if user == nil && isUserID {
		return nil, handler.ErrNotFound("unknown_user", "User not found. They need to log in to Kite once, or invite them by their Discord username instead.")
	}

	var userID null.String
	if user != nil {
		if user.ID == c.App.OwnerUserID {
			return nil, handler.ErrBadRequest("cannot_add_owner", "Cannot add owner as collaborator")
		}

		_, err = h.appStore.Collaborator(c.Context(), c.App.ID, user.ID)
		if err == nil {
			return nil, handler.ErrBadRequest("already_collaborator", "This user is already a collaborator")
		} else if !errors.Is(err, store.ErrNotFound) {
			return nil, err
		}

		userID = null.StringFrom(user.ID)
		// Invites are keyed by username, so one by ID replaces one by name.
		discordUsername = strings.ToLower(user.DiscordUsername)
	}

	invite, err := h.appStore.UpsertCollaboratorInvite(c.Context(), &model.AppCollaboratorInvite{
		AppID:           c.App.ID,
		DiscordUsername: discordUsername,
		UserID:          userID,
		Role:            role,
		InvitedByUserID: null.StringFrom(c.Session.UserID),
		CreatedAt:       now,
		ExpiresAt:       now.Add(collaboratorInviteTTL),
	})
	if err != nil {
		return nil, err
	}

	c.SetAuditTarget(handler.AuditTarget{
		ID:   discordUsername,
		Name: "@" + discordUsername + " as " + string(role),
	})

	return &wire.AppCollaboratorCreateResponse{
		Invite: wire.CollaboratorInviteToWire(invite),
	}, nil
}

// HandleAppCollaboratorUpdate changes the role of a collaborator.
func (h *AppHandler) HandleAppCollaboratorUpdate(c *handler.Context, req wire.AppCollaboratorUpdateRequest) (*wire.AppCollaboratorUpdateResponse, error) {
	if !c.UserAppRole.CanChangeRoles() {
		return nil, handler.ErrForbidden("missing_permissions", "Only the owner of the app can change the roles of collaborators")
	}

	collaborator, err := h.appStore.Collaborator(c.Context(), c.App.ID, c.Param("userID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_collaborator", "Collaborator not found")
		}
		return nil, err
	}

	user := collaborator.User
	collaborator.Role = model.AppCollaboratorRole(req.Role)
	collaborator.UpdatedAt = time.Now().UTC()

	updated, err := h.appStore.UpdateCollaborator(c.Context(), collaborator)
	if err != nil {
		return nil, err
	}
	updated.User = user

	c.SetAuditTarget(handler.AuditTarget{
		Name: collaboratorAuditName(user) + " to " + req.Role,
	})

	return wire.CollaboratorToWire(updated), nil
}

func (h *AppHandler) HandleAppCollaboratorDelete(c *handler.Context) (*wire.AppCollaboratorDeleteResponse, error) {
	userID := c.Param("userID")

	collaborator, err := h.appStore.Collaborator(c.Context(), c.App.ID, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_collaborator", "Collaborator not found")
		}
		return nil, err
	}

	leaving := userID == c.Session.UserID
	// Collaborators can always remove themselves to leave the app. Admins
	// can only remove developers, not themselves or other admins from the
	// list of collaborators.
	if !leaving && !c.UserAppRole.CanRemoveRole(collaborator.Role) {
		if c.UserAppRole == model.AppCollaboratorRoleAdmin {
			return nil, handler.ErrForbidden("missing_permissions", "Admins can't remove other admins, only the owner of the app can")
		}
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to remove collaborators from this app")
	}

	err = h.appStore.DeleteCollaborator(c.Context(), c.App.ID, userID)
	if err != nil {
		return nil, err
	}

	target := handler.AuditTarget{Name: collaboratorAuditName(collaborator.User)}
	if leaving {
		target.Action = "collaborator.leave"
	}
	c.SetAuditTarget(target)

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
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to revoke invites of this app")
	}

	discordUsername := c.Param("discordUsername")
	invite, err := h.appStore.CollaboratorInvite(c.Context(), c.App.ID, discordUsername)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_invite", "Invite not found")
		}
		return nil, err
	}

	if !c.UserAppRole.CanRemoveRole(invite.Role) {
		return nil, handler.ErrForbidden("missing_permissions", "Only the owner of the app can revoke invites of admins")
	}

	err = h.appStore.DeleteCollaboratorInvite(c.Context(), c.App.ID, discordUsername)
	if err != nil {
		return nil, err
	}

	c.SetAuditTarget(handler.AuditTarget{Name: "@" + discordUsername})

	return &wire.AppCollaboratorInviteDeleteResponse{}, nil
}

const (
	defaultAuditLogLimit = 50
	maxAuditLogLimit     = 100
)

func (h *AppHandler) HandleAppAuditLogList(c *handler.Context) (*wire.AppAuditLogListResponse, error) {
	if !c.UserAppRole.CanViewAuditLog() {
		return nil, handler.ErrForbidden("missing_permissions", "Only admins and the owner of the app can see its audit log")
	}

	var beforeID int64
	if before := c.Query("before"); before != "" {
		id, err := strconv.ParseInt(before, 10, 64)
		if err != nil || id <= 0 {
			return nil, handler.ErrBadRequest("invalid_before", "before must be the ID of an audit log entry")
		}
		beforeID = id
	}

	limit := defaultAuditLogLimit
	if l := c.Query("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n < 1 || n > maxAuditLogLimit {
			return nil, handler.ErrBadRequest("invalid_limit", fmt.Sprintf("limit must be between 1 and %d", maxAuditLogLimit))
		}
		limit = n
	}

	entries, err := h.appStore.AppAuditLogEntries(c.Context(), c.App.ID, beforeID, limit)
	if err != nil {
		return nil, err
	}

	res := make(wire.AppAuditLogListResponse, len(entries))
	for i, entry := range entries {
		res[i] = wire.AppAuditLogEntryToWire(entry)
	}

	return &res, nil
}

func collaboratorAuditName(user *model.User) string {
	if user == nil {
		return ""
	}
	return "@" + user.DiscordUsername
}
