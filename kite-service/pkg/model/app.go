package model

import (
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"gopkg.in/guregu/null.v4"
)

type App struct {
	ID                  distype.Snowflake
	OwnerUserID         distype.Snowflake
	Token               string
	TokenInvalid        bool
	PublicKey           string
	UserID              distype.Snowflake
	UserName            string
	UserDiscriminator   string
	UserAvatar          null.String
	UserBanner          null.String
	UserBio             null.String
	StatusType          string
	StatusActivityType  null.Int
	StatusActivityName  null.String
	StatusActivityState null.String
	StatusActivityUrl   null.String
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
