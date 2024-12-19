package wire

import (
	"time"

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
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
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
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
