package wire

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID            distype.Snowflake `json:"id"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	Discriminator null.String       `json:"discriminator"`
	GlobalName    null.String       `json:"global_name"`
	Avatar        null.String       `json:"avatar"`
	PublicFlags   int               `json:"public_flags"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type UserGetResponse APIResponse[User]

func UserToWire(user *model.User) User {
	return User{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Discriminator: user.Discriminator,
		GlobalName:    user.GlobalName,
		Avatar:        user.Avatar,
		PublicFlags:   user.PublicFlags,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
