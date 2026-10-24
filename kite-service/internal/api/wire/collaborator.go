package wire

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

var (
	discordUserIDRegex   = regexp.MustCompile(`^[0-9]{15,21}$`)
	discordUsernameRegex = regexp.MustCompile(`^@?[A-Za-z0-9_.]{2,32}$`)
)

type AppCollaborator struct {
	User      User      `json:"user"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AppCollaboratorListResponse = []*AppCollaborator

type AppCollaboratorInvite struct {
	DiscordUsername string    `json:"discord_username"`
	Role            string    `json:"role"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

type AppCollaboratorCreateRequest struct {
	DiscordUserID   string `json:"discord_user_id"`
	DiscordUsername string `json:"discord_username"`
	Role            string `json:"role"`
}

func (req AppCollaboratorCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.DiscordUserID,
			validation.When(req.DiscordUsername == "", validation.Required),
			validation.Match(discordUserIDRegex),
		),
		validation.Field(&req.DiscordUsername,
			validation.When(req.DiscordUserID == "", validation.Required),
			validation.Match(discordUsernameRegex),
		),
		validation.Field(&req.Role, validation.Required, validation.In(string(model.AppCollaboratorRoleAdmin))),
	)
}

// AppCollaboratorCreateResponse has the invite instead of the collaborator when
// the user hasn't logged in to Kite yet.
type AppCollaboratorCreateResponse struct {
	Collaborator *AppCollaborator       `json:"collaborator"`
	Invite       *AppCollaboratorInvite `json:"invite"`
}

type AppCollaboratorDeleteResponse = Empty

type AppCollaboratorInviteListResponse = []*AppCollaboratorInvite

type AppCollaboratorInviteDeleteResponse = Empty

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

func CollaboratorInviteToWire(invite *model.AppCollaboratorInvite) *AppCollaboratorInvite {
	if invite == nil {
		return nil
	}

	return &AppCollaboratorInvite{
		DiscordUsername: invite.DiscordUsername,
		Role:            string(invite.Role),
		CreatedAt:       invite.CreatedAt,
		ExpiresAt:       invite.ExpiresAt,
	}
}
