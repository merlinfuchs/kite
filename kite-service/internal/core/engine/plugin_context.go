package engine

import (
	"context"
	"encoding/json"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type pluginContext struct {
	context.Context
	*pluginValueProvider
	client *api.Client
}

func newPluginContext(
	ctx context.Context,
	store store.PluginValueStore,
	client *api.Client,
	appID string,
	pluginID string,
) *pluginContext {
	return &pluginContext{
		Context:             ctx,
		pluginValueProvider: newPluginValueProvider(ctx, store, appID, pluginID),
		client:              client,
	}
}

func (c *pluginContext) Client() *api.Client {
	return c.client
}

type pluginValueProvider struct {
	ctx   context.Context
	store store.PluginValueStore

	appID    string
	pluginID string
}

func newPluginValueProvider(
	ctx context.Context,
	store store.PluginValueStore,
	appID string,
	pluginID string,
) *pluginValueProvider {
	return &pluginValueProvider{
		ctx:      ctx,
		store:    store,
		appID:    appID,
		pluginID: pluginID,
	}
}

func (p *pluginValueProvider) SetKey(key string, value json.RawMessage) error {
	return p.store.SetPluginValue(p.ctx, p.appID, p.pluginID, key, value)
}

func (p *pluginValueProvider) GetKey(key string) (json.RawMessage, error) {
	return p.store.GetPluginValue(p.ctx, p.appID, p.pluginID, key)
}

func (p *pluginValueProvider) IncreaseKey(key string, amount int) (int, error) {
	return p.store.IncreasePluginValue(p.ctx, p.appID, p.pluginID, key, amount)
}

func (p *pluginValueProvider) DecreaseKey(key string, amount int) (int, error) {
	return p.store.DecreasePluginValue(p.ctx, p.appID, p.pluginID, key, amount)
}

func (p *pluginValueProvider) DeleteKey(key string) (json.RawMessage, error) {
	return p.store.DeletePluginValue(p.ctx, p.appID, p.pluginID, key)
}
