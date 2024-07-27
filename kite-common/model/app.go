package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type App struct {
	ID            string
	Name          string
	Description   null.String
	OwnerUserID   string
	CreatorUserID string
	DiscordToken  string
	DiscordID     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AppCredentials struct {
	DiscordID    string
	DiscordToken string
}
