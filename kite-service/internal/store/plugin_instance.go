package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type PluginInstanceStore interface {
	PluginInstancesByApp(ctx context.Context, appID string) ([]*model.PluginInstance, error)
	PluginInstance(ctx context.Context, appID string, pluginID string) (*model.PluginInstance, error)
	CreatePluginInstance(ctx context.Context, pluginInstance *model.PluginInstance) (*model.PluginInstance, error)
	UpdatePluginInstance(ctx context.Context, pluginInstance *model.PluginInstance) (*model.PluginInstance, error)
	UpdatePluginInstancesLastDeployedAt(ctx context.Context, appID string, lastDeployedAt time.Time) error
	EnabledPluginInstancesUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.PluginInstance, error)
	EnabledPluginInstanceIDs(ctx context.Context) ([]string, error)
	DeletePluginInstance(ctx context.Context, appID string, pluginID string) error
	DinstinctAppIDsWithUndeployedPluginInstances(ctx context.Context) ([]string, error)
}
