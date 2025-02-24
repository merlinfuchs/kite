package wire

import (
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type AppCollaborator struct {
	User      User      `json:"user"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AppCollaboratorListResponse = []*AppCollaborator

type AppCollaboratorCreateRequest struct {
	DiscordUserID string `json:"discord_user_id"`
	Role          string `json:"role"`
}

type AppCollaboratorCreateResponse = AppCollaborator

type AppCollaboratorDeleteResponse = Empty

func CollaboratorToWire(collaborator *model.AppCollaborator) *AppCollaborator {
	if collaborator == nil {
		return nil
	}

	var user User
	if collaborator.User != nil {
		user = *UserToWire(collaborator.User, true)
	} else {
		user.ID = collaborator.UserID
	}

	return &AppCollaborator{
		User:      user,
		Role:      string(collaborator.Role),
		CreatedAt: collaborator.CreatedAt,
		UpdatedAt: collaborator.UpdatedAt,
	}
}
