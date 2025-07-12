package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
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
	Data       thing.Any
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type VariableValueOperation = provider.VariableOperation
