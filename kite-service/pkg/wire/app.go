package wire

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"gopkg.in/guregu/null.v4"
)

type App struct {
	ID                  distype.Snowflake `json:"id"`
	OwnerUserID         distype.Snowflake `json:"owner_user_id"`
	TokenInvalid        bool              `json:"token_invalid"`
	PublicKey           string            `json:"public_key"`
	UserID              distype.Snowflake `json:"user_id"`
	UserName            string            `json:"user_name"`
	UserDiscriminator   string            `json:"user_discriminator"`
	UserAvatar          null.String       `json:"user_avatar"`
	UserBanner          null.String       `json:"user_banner"`
	UserBio             null.String       `json:"user_bio"`
	StatusType          string            `json:"status_type"`
	StatusActivityType  null.Int          `json:"status_activity_type"`
	StatusActivityName  null.String       `json:"status_activity_name"`
	StatusActivityState null.String       `json:"status_activity_state"`
	StatusActivityUrl   null.String       `json:"status_activity_url"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type AppCreateRequest struct {
	Token string `json:"token" validate:"required"`
}

type AppCreateResponse APIResponse[App]

type AppListResponse APIResponse[[]App]

type AppGetResponse APIResponse[App]

type AppTokenUpdateRequest struct {
	Token string `json:"token" validate:"required"`
}

type AppTokenUpdateResponse APIResponse[App]

type AppStatusUpdateRequest struct {
	StatusType    string      `json:"status_type"`
	ActivityType  null.Int    `json:"activity_type"`
	ActivityName  null.String `json:"activity_name"`
	ActivityState null.String `json:"activity_state"`
	ActivityUrl   null.String `json:"activity_url"`
}

type AppStatusUpdateResponse APIResponse[App]

type AppUserUpdateRequest struct {
	UserName   string      `json:"user_name"`
	UserAvatar null.String `json:"user_avatar"`
	UserBanner null.String `json:"user_banner"`
	UserBio    null.String `json:"user_bio"`
}

type AppUserUpdateResponse APIResponse[App]

func AppToWire(a *model.App) App {
	return App{
		ID:                  a.ID,
		OwnerUserID:         a.OwnerUserID,
		TokenInvalid:        a.TokenInvalid,
		PublicKey:           a.PublicKey,
		UserID:              a.UserID,
		UserName:            a.UserName,
		UserDiscriminator:   a.UserDiscriminator,
		UserAvatar:          a.UserAvatar,
		UserBanner:          a.UserBanner,
		UserBio:             a.UserBio,
		StatusType:          a.StatusType,
		StatusActivityType:  a.StatusActivityType,
		StatusActivityName:  a.StatusActivityName,
		StatusActivityState: a.StatusActivityState,
		StatusActivityUrl:   a.StatusActivityUrl,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
}
