package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Session struct {
	KeyHash   string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time

	// KiteBanReason is set when Kite banned the user, see UserBanFeatureKite.
	// It's only loaded by the lookup of a request's session.
	KiteBanReason null.String
}
