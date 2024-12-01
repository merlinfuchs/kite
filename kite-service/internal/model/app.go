package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type App struct {
	ID             string
	Name           string
	Description    null.String
	Enabled        bool
	DisabledReason null.String
	OwnerUserID    string
	CreatorUserID  string
	DiscordToken   string
	DiscordID      string
	DiscordStatus  *AppDiscordStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AppCredentials struct {
	DiscordID    string
	DiscordToken string
}

type AppDiscordStatus struct {
	Status        string `json:"status,omitempty"`
	ActivityType  int    `json:"activity_type,omitempty"`
	ActivityName  string `json:"activity_name,omitempty"`
	ActivityState string `json:"activity_state,omitempty"`
	ActivityURL   string `json:"activity_url,omitempty"`
}

func (s *AppDiscordStatus) Equals(other *AppDiscordStatus) bool {
	if s == nil && other == nil {
		return true
	}

	if s == nil || other == nil {
		return false
	}

	return s.Status == other.Status &&
		s.ActivityType == other.ActivityType &&
		s.ActivityName == other.ActivityName &&
		s.ActivityState == other.ActivityState &&
		s.ActivityURL == other.ActivityURL
}
