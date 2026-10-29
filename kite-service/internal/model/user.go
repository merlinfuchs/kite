package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID              string
	Email           string
	DisplayName     string
	DiscordID       string
	DiscordUsername string
	DiscordAvatar   null.String
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserBanFeature is a feature Kite can ban users from.
type UserBanFeature string

const (
	// UserBanFeatureMarketplace keeps the user from submitting bundles to the
	// marketplace, and new versions of their bundles aren't reviewed. What
	// they listed before stays listed.
	UserBanFeatureMarketplace UserBanFeature = "marketplace"
	// UserBanFeatureKite keeps the user from using Kite at all: every request
	// with their login or API tokens is refused, and their apps are stopped
	// until they're unbanned. It includes the marketplace ban.
	UserBanFeatureKite UserBanFeature = "kite"
)

// UserBan bans a user from a feature.
type UserBan struct {
	UserID  string
	Feature UserBanFeature
	// Reason is shown to the user.
	Reason         string
	BannedByUserID null.String
	CreatedAt      time.Time
	// User is set when bans are shown to the admins.
	User *User
}
