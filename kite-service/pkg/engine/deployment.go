package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/go-types/logmodel"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
	"github.com/tetratelabs/wazero"
)

type PluginDeployment struct {
	ID               string
	wasm             []byte
	manifest         plugin.Manifest
	config           plugin.PluginConfig
	env              plugin.HostEnvironment
	compilationCache wazero.CompilationCache

	pluginPool *pool.ObjectPool
}

func (pd *PluginDeployment) Manifest() plugin.Manifest {
	return pd.manifest
}

func (pd *PluginDeployment) Config() plugin.PluginConfig {
	return pd.config
}

func NewDeployment(
	id string,
	env plugin.HostEnvironment,
	compilationCache wazero.CompilationCache,
	wasm []byte,
	manifest plugin.Manifest,
	config plugin.PluginConfig,
) *PluginDeployment {
	dp := &PluginDeployment{
		ID:               id,
		env:              env,
		wasm:             wasm,
		manifest:         manifest,
		config:           config,
		compilationCache: compilationCache,
	}

	factory := pool.NewPooledObjectFactorySimple(dp.pluginFactory)
	dp.pluginPool = pool.NewObjectPool(context.Background(), factory, &pool.ObjectPoolConfig{
		LIFO:                     true,
		MaxTotal:                 4,
		MaxIdle:                  4,
		MinIdle:                  0,
		SoftMinEvictableIdleTime: 60 * time.Second,
		TimeBetweenEvictionRuns:  10 * time.Second,
	})
	dp.pluginPool.StartEvictor()

	return dp
}

func (dp *PluginDeployment) pluginFactory(ctx context.Context) (interface{}, error) {
	p, err := plugin.New(ctx, dp.wasm, dp.manifest, dp.config, dp.env, dp.compilationCache)
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

	res, err := plugin.Handle(ctx, event)

	pd.env.TrackEventHandled(ctx, string(event.Type), err == nil, res.TotalDuration, res.ExecutionDuration)

	if err != nil {
		// TODO: think about other error types that should invalidate the plugin (e.g. panic / wasm trap)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			if err := pd.InvalidatePlugin(ctx, plugin); err != nil {
				slog.With(logattr.Error(err)).Error("failed to invalidate plugin")
			}
		}

		pd.env.Log(ctx, logmodel.LogLevelError, err.Error())
	} else {
		if err := pd.ReturnPlugin(ctx, plugin); err != nil {
			slog.With(logattr.Error(err)).Error("failed to return plugin")
		}
	}

	return nil
}
