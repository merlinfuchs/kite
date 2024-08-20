package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Variable struct {
	ID          string
	Scope       string
	Name        string
	Type        string
	AppID       string
	ModuleID    null.String
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TotalValues null.Int
}
