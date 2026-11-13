package wire

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

var (
	DiscordUserIDRegex   = regexp.MustCompile(`^[0-9]{15,21}$`)
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
	DiscordUsername string `json:"discord_username"`
	// UserID is set when the invited user already uses Kite.
	UserID    null.String `json:"user_id"`
	Role      string      `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// invitableCollaboratorRoles are the roles collaborators can have, the app
// has one owner.
var invitableCollaboratorRoles = []any{
	string(model.AppCollaboratorRoleAdmin),
	string(model.AppCollaboratorRoleDeveloper),
}

type AppCollaboratorCreateRequest struct {
	// DiscordUser is the user's Discord ID or username.
	DiscordUser string `json:"discord_user"`
	Role        string `json:"role"`
}

func (req AppCollaboratorCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.DiscordUser,
			validation.Required,
			validation.By(func(value interface{}) error {
				v := value.(string)
				if DiscordUserIDRegex.MatchString(v) || discordUsernameRegex.MatchString(v) {
					return nil
				}
				return errors.New("must be a Discord username or user ID")
			}),
		),
		validation.Field(&req.Role, validation.Required, validation.In(invitableCollaboratorRoles...)),
	)
}

// AppCollaboratorCreateResponse is the invite, the user becomes a
// collaborator once they accept it.
type AppCollaboratorCreateResponse struct {
	Invite *AppCollaboratorInvite `json:"invite"`
}

type AppCollaboratorUpdateRequest struct {
	Role string `json:"role"`
}

func (req AppCollaboratorUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Role, validation.Required, validation.In(invitableCollaboratorRoles...)),
	)
}

type AppCollaboratorUpdateResponse = AppCollaborator

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
		UserID:          invite.UserID,
		Role:            string(invite.Role),
		CreatedAt:       invite.CreatedAt,
		ExpiresAt:       invite.ExpiresAt,
	}
}

// UserCollaboratorInvite is an invite of the current user to another app.
type UserCollaboratorInvite struct {
	App       UserCollaboratorInviteApp `json:"app"`
	Role      string                    `json:"role"`
	InvitedBy *User                     `json:"invited_by"`
	CreatedAt time.Time                 `json:"created_at"`
	ExpiresAt time.Time                 `json:"expires_at"`
}

// UserCollaboratorInviteApp is what the invited user sees of the app before
// they accept.
type UserCollaboratorInviteApp struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description null.String `json:"description"`
	DiscordID   string      `json:"discord_id"`
}

type UserCollaboratorInviteListResponse = []*UserCollaboratorInvite

type UserCollaboratorInviteAcceptResponse = AppCollaborator

type UserCollaboratorInviteDeclineResponse = Empty

func UserCollaboratorInviteToWire(invite *model.AppCollaboratorInvite) *UserCollaboratorInvite {
	if invite == nil || invite.App == nil {
		return nil
	}

	return &UserCollaboratorInvite{
		App: UserCollaboratorInviteApp{
			ID:          invite.App.ID,
			Name:        invite.App.Name,
			Description: invite.App.Description,
			DiscordID:   invite.App.DiscordID,
		},
		Role:      string(invite.Role),
		InvitedBy: UserToWire(invite.InvitedBy, false),
		CreatedAt: invite.CreatedAt,
		ExpiresAt: invite.ExpiresAt,
	}
}

// AppAuditLogEntry is something a collaborator did in the app.
type AppAuditLogEntry struct {
	ID string `json:"id"`
	// User is null once they deleted their account.
	User       *User  `json:"user"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	TargetName string `json:"target_name"`
	// ViaAPIToken is whether the change was made with an API token, like
	// by an MCP client.
	ViaAPIToken bool      `json:"via_api_token"`
	CreatedAt   time.Time `json:"created_at"`
}

type AppAuditLogListResponse = []*AppAuditLogEntry

func AppAuditLogEntryToWire(entry *model.AppAuditLogEntry) *AppAuditLogEntry {
	if entry == nil {
		return nil
	}

	return &AppAuditLogEntry{
		ID:          strconv.FormatInt(entry.ID, 10),
		User:        UserToWire(entry.User, false),
		Action:      entry.Action,
		TargetType:  entry.TargetType,
		TargetID:    entry.TargetID,
		TargetName:  entry.TargetName,
		ViaAPIToken: entry.APITokenID.Valid,
		CreatedAt:   entry.CreatedAt,
	}
}
