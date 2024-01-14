package engine

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
)

type PluginDeployment struct {
	wasm     []byte
	manifest plugin.Manifest
	config   plugin.PluginConfig
	env      plugin.HostEnvironment

	pluginPool *pool.ObjectPool
}

func (pd *PluginDeployment) Manifest() plugin.Manifest {
	return pd.manifest
}

func (pd *PluginDeployment) Config() plugin.PluginConfig {
	return pd.config
}

func NewDeployment(wasm []byte, manifest plugin.Manifest, config plugin.PluginConfig, env plugin.HostEnvironment) *PluginDeployment {
	dp := &PluginDeployment{
		wasm:     wasm,
		manifest: manifest,
		config:   config,
		env:      env,
	}

	factory := pool.NewPooledObjectFactorySimple(dp.pluginFactory)
	dp.pluginPool = pool.NewObjectPoolWithDefaultConfig(context.Background(), factory)

	return dp
}

func (dp *PluginDeployment) pluginFactory(ctx context.Context) (interface{}, error) {
	p, err := plugin.New(ctx, dp.wasm, dp.manifest, dp.config, dp.env)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (pd *PluginDeployment) BorrowPlugin(ctx context.Context) (*plugin.Plugin, error) {
	obj, err := pd.pluginPool.BorrowObject(ctx)
	if err != nil {
		return nil, err
	}
	return obj.(*plugin.Plugin), nil
}

func (pd *PluginDeployment) ReturnPlugin(ctx context.Context, p *plugin.Plugin) error {
	return pd.pluginPool.ReturnObject(ctx, p)
}

func (pd *PluginDeployment) InvalidatePlugin(ctx context.Context, p *plugin.Plugin) error {
	return pd.pluginPool.InvalidateObject(ctx, p)
}

func (pd *PluginDeployment) HandleEvent(ctx context.Context, event *event.Event) error {
	if !slices.Contains(pd.Manifest().Events, string(event.Type)) {
		return nil
	}

	plugin, err := pd.BorrowPlugin(ctx)
	if err != nil {
		return fmt.Errorf("failed to borrow plugin: %w", err)
	}

	err = plugin.Handle(ctx, event)

	if err := pd.ReturnPlugin(ctx, plugin); err != nil {
		slog.With(logattr.Error(err)).Error("failed to return plugin")
	}

	if err != nil {
		return err
	}

	return nil
}
