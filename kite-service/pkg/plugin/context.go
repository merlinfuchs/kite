package plugin

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/diamondburned/arikawa/v3/state"
)

type Context interface {
	context.Context
	ValueStore

	Discord() *state.State
}

var ErrValueNotFound = errors.New("value not found")
var ErrValueWrongType = errors.New("value is of wrong type")

type ValueStore interface {
	SetKey(key string, value json.RawMessage) error
	GetKey(key string) (json.RawMessage, error)
	IncreaseKey(key string, amount int) (int, error)
	DecreaseKey(key string, amount int) (int, error)
	DeleteKey(key string) (json.RawMessage, error)
}
