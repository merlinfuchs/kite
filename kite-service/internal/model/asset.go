package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Asset struct {
	ID            string
	AppID         string
	ModuleID      null.String
	CreatorUserID string
	Name          string
	ContentType   string
	ContentHash   string
	ContentSize   int
	Content       []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     null.Time
}
