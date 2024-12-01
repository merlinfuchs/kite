package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type APIKey struct {
	ID            string
	Type          APIKeyType
	Name          string
	Key           string
	KeyHash       string
	AppID         string
	CreatorUserID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     null.Time
}

type APIKeyType string

const (
	APIKeyTypeIFTTT APIKeyType = "ifttt"
)
