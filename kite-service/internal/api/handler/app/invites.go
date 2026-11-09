package app

import (
	"errors"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// HandleUserInviteList lists the invites of the current user to other apps.
func (h *AppHandler) HandleUserInviteList(c *handler.Context) (*wire.UserCollaboratorInviteListResponse, error) {
	user, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, err
	}

	invites, err := h.appStore.CollaboratorInvitesByUser(c.Context(), user.ID, user.DiscordUsername, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	res := make(wire.UserCollaboratorInviteListResponse, 0, len(invites))
	for _, invite := range invites {
		if invite.InvitedByUserID.Valid {
			inviter, err := h.userStore.User(c.Context(), invite.InvitedByUserID.String)
			if err == nil {
				invite.InvitedBy = inviter
			} else if !errors.Is(err, store.ErrNotFound) {
				return nil, err
			}
		}
		res = append(res, wire.UserCollaboratorInviteToWire(invite))
	}

	return &res, nil
}

// HandleUserInviteAccept makes the current user a collaborator of the app that
// invited them.
func (h *AppHandler) HandleUserInviteAccept(c *handler.Context) (*wire.UserCollaboratorInviteAcceptResponse, error) {
	user, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, err
	}

	appID := c.Param("appID")
	now := time.Now().UTC()

	collaborator, err := h.appStore.AcceptCollaboratorInvite(c.Context(), appID, user.ID, user.DiscordUsername, now)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_invite", "Invite not found, it might have expired or been revoked")
		}
		return nil, err
	}
	collaborator.User = user

	h.createAuditLogEntry(c, &model.AppAuditLogEntry{
		AppID:      appID,
		UserID:     null.StringFrom(user.ID),
		Action:     "collaborator.join",
		TargetType: "collaborator",
		TargetID:   user.ID,
		TargetName: "as " + string(collaborator.Role),
		CreatedAt:  now,
	})

	return wire.CollaboratorToWire(collaborator), nil
}

// HandleUserInviteDecline deletes the invite of the current user to the app.
func (h *AppHandler) HandleUserInviteDecline(c *handler.Context) (*wire.UserCollaboratorInviteDeclineResponse, error) {
	user, err := h.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, err
	}

	appID := c.Param("appID")

	err = h.appStore.DeclineCollaboratorInvite(c.Context(), appID, user.ID, user.DiscordUsername)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_invite", "Invite not found")
		}
		return nil, err
	}

	h.createAuditLogEntry(c, &model.AppAuditLogEntry{
		AppID:      appID,
		UserID:     null.StringFrom(user.ID),
		Action:     "collaborator.decline",
		TargetType: "collaborator",
		TargetID:   user.ID,
		CreatedAt:  time.Now().UTC(),
	})

	return &wire.UserCollaboratorInviteDeclineResponse{}, nil
}

func (h *AppHandler) createAuditLogEntry(c *handler.Context, entry *model.AppAuditLogEntry) {
	if err := h.appStore.CreateAppAuditLogEntry(c.Context(), entry); err != nil {
		slog.Error(
			"Failed to create audit log entry",
			slog.String("app_id", entry.AppID),
			slog.String("action", entry.Action),
			slog.String("error", err.Error()),
		)
	}
}
