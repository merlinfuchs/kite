package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type PluginValueStore interface {
	SetPluginValue(ctx context.Context, value model.PluginValue) error
	UpdatePluginValue(ctx context.Context, operation model.PluginValueOperation, value model.PluginValue) (*model.PluginValue, error)
	GetPluginValue(ctx context.Context, pluginInstanceID, key string) (*model.PluginValue, error)
	DeletePluginValue(ctx context.Context, pluginInstanceID, key string) error
}
