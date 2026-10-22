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
	SignupSource    SignupSource
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SignupSource is where the user first came from before signing up, as
// recorded by the website. It's only stored when the user is created.
type SignupSource struct {
	Referrer    null.String
	UTMSource   null.String
	UTMMedium   null.String
	UTMCampaign null.String
	LandingPage null.String
}
