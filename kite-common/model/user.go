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
