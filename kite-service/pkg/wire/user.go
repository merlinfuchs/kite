package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type User struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	GlobalName    string    `json:"global_name"`
	Avatar        string    `json:"avatar"`
	PublicFlags   int       `json:"public_flags"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UserGetResponse APIResponse[User]

func UserToWire(user *model.User) User {
	return User{
		ID:            user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		GlobalName:    user.GlobalName,
		Avatar:        user.Avatar,
		PublicFlags:   user.PublicFlags,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
