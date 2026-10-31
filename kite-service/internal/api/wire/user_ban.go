package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// UserBan bans a user from a feature, for the admins.
type UserBan struct {
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
	// Feature is what they're banned from, marketplace or kite.
	Feature   string    `json:"feature"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type UserBanListResponse = []*UserBan

// UserBanCreateRequest bans a user from a feature, replacing their ban from
// it.
type UserBanCreateRequest struct {
	// UserID is the Kite user to ban. Set either it or DiscordID.
	UserID string `json:"user_id,omitempty"`
	// DiscordID is the Discord ID of the user or of one of their bots.
	DiscordID string `json:"discord_id,omitempty"`
	Feature   string `json:"feature"`
	// Reason is shown to the user.
	Reason string `json:"reason"`
}

func (req UserBanCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.When(req.DiscordID == "", validation.Required), validation.When(req.DiscordID != "", validation.Empty)),
		validation.Field(&req.DiscordID, is.Digit),
		validation.Field(&req.Feature, validation.Required, validation.In(string(model.UserBanFeatureMarketplace), string(model.UserBanFeatureKite))),
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
