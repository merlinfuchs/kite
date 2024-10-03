package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Variable struct {
	ID          string
	Name        string
	Scoped      bool
	AppID       string
	ModuleID    null.String
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TotalValues null.Int
}

type VariableValue struct {
	ID         uint64
	VariableID string
	Scope      null.String
	Value      []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
