package wire

import (
	"fmt"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID              string      `json:"id"`
	Email           null.String `json:"email"`
	DisplayName     string      `json:"display_name"`
	DiscordID       string      `json:"discord_id"`
	DiscordUsername string      `json:"discord_username"`
	DiscordAvatar   null.String `json:"discord_avatar"`
	// Bio is the text they show on their creator page in the marketplace.
	Bio string `json:"bio"`
	// IsAdmin is whether the current user can review bundles, false for
	// other users.
	IsAdmin bool `json:"is_admin"`
	// Badges are the badges of the current user, empty for other users.
	Badges    []UserBadge `json:"badges,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type UserGetResponse = User

func UserToWire(user *model.User, withEmail bool) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:              user.ID,
		Email:           null.NewString(user.Email, withEmail),
		DisplayName:     user.DisplayName,
		DiscordID:       user.DiscordID,
		DiscordUsername: user.DiscordUsername,
		DiscordAvatar:   user.DiscordAvatar,
		Bio:             user.Bio,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// UserUpdateRequest changes the profile of the current user.
type UserUpdateRequest struct {
	// Bio is the text they show on their creator page in the marketplace, up
	// to MaxUserBioLength characters on MaxUserBioLines lines. Empty removes
	// it.
	Bio string `json:"bio"`
}

func (req UserUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Bio,
			validation.RuneLength(0, model.MaxUserBioLength),
			validation.By(func(value any) error {
				if strings.Count(value.(string), "\n")+1 > model.MaxUserBioLines {
					return validation.NewError("validation_bio_lines", fmt.Sprintf("must have at most %d lines", model.MaxUserBioLines))
				}
				return nil
			}),
		),
	)
}

// Sanitize trims the bio and drops empty lines that would only add space,
// so the limits count what is shown.
func (req *UserUpdateRequest) Sanitize() {
	bio := strings.ReplaceAll(req.Bio, "\r\n", "\n")
	bio = strings.ReplaceAll(bio, "\r", "\n")
	lines := strings.Split(bio, "\n")
	kept := lines[:0]
	for i, line := range lines {
		line = strings.TrimRight(line, " \t")
		// One empty line can separate paragraphs, more can't.
		if line == "" && i > 0 && len(kept) > 0 && kept[len(kept)-1] == "" {
			continue
		}
		kept = append(kept, line)
	}
	req.Bio = strings.TrimSpace(strings.Join(kept, "\n"))
}

type UserUpdateResponse = User
