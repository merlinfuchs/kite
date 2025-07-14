package plugin

import (
	"context"
	"errors"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type Context interface {
	context.Context

	provider.ValueProvider

	Discord() provider.DiscordProvider
}

var ErrValueNotFound = errors.New("value not found")
var ErrValueWrongType = errors.New("value is of wrong type")
