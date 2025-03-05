package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type PluginInstanceStore interface {
	EnabledPluginInstanceIDs(ctx context.Context) (map[string][]string, error)
	EnabledPluginInstancesUpdatedSince(ctx context.Context, lastUpdate time.Time) ([]*model.PluginInstance, error)
	PluginInstance(ctx context.Context, appID string, pluginID string) (*model.PluginInstance, error)
	UpsertPluginInstance(ctx context.Context, pluginInstance model.PluginInstance) (*model.PluginInstance, error)
	UpdatePluginInstancesCommandsDeployedAt(ctx context.Context, appID string, commandsDeployedAt time.Time) error
}

type PluginValueStore interface {
	SetPluginValue(ctx context.Context, appID string, pluginID string, key string, value json.RawMessage) error
	GetPluginValue(ctx context.Context, appID string, pluginID string, key string) (json.RawMessage, error)
	DeletePluginValue(ctx context.Context, appID string, pluginID string, key string) (json.RawMessage, error)
	IncreasePluginValue(ctx context.Context, appID string, pluginID string, key string, amount int) (int, error)
	DecreasePluginValue(ctx context.Context, appID string, pluginID string, key string, amount int) (int, error)
}
