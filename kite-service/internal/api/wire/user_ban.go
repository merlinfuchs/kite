package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// UserBan bans a user from a feature, for the admins.
type UserBan struct {
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
	// Feature is what they're banned from, only marketplace for now.
	Feature   string    `json:"feature"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type UserBanListResponse = []*UserBan

// UserBanCreateRequest bans a user from a feature, replacing their ban from
// it.
type UserBanCreateRequest struct {
	UserID  string `json:"user_id"`
	Feature string `json:"feature"`
	// Reason is shown to the user.
	Reason string `json:"reason"`
}

func (req UserBanCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required),
		validation.Field(&req.Feature, validation.Required, validation.In(string(model.UserBanFeatureMarketplace))),
		validation.Field(&req.Reason, validation.Required, validation.Length(1, 1000)),
	)
}

type UserBanCreateResponse = UserBan

type UserBanDeleteResponse struct{}

func UserBanToWire(ban *model.UserBan) *UserBan {
	return &UserBan{
		UserID:    ban.UserID,
		User:      UserToWire(ban.User, false),
		Feature:   string(ban.Feature),
		Reason:    ban.Reason,
		CreatedAt: ban.CreatedAt,
	}
}
