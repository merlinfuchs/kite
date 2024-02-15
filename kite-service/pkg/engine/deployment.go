package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-sdk-go/log"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/module"
)

type Deployment struct {
	ID       string
	wasm     []byte
	manifest manifest.Manifest
	config   module.ModuleConfig
	env      module.HostEnvironment

	pluginPool *pool.ObjectPool
}

func (d *Deployment) Manifest() manifest.Manifest {
	return d.manifest
}

func (d *Deployment) Config() module.ModuleConfig {
	return d.config
}

func NewDeployment(
	id string,
	env module.HostEnvironment,
	wasm []byte,
	manifest manifest.Manifest,
	config module.ModuleConfig,
) *Deployment {
	dp := &Deployment{
		ID:       id,
		env:      env,
		wasm:     wasm,
		manifest: manifest,
		config:   config,
	}

	factory := pool.NewPooledObjectFactorySimple(dp.pluginFactory)
	dp.pluginPool = pool.NewObjectPool(context.Background(), factory, &pool.ObjectPoolConfig{
		LIFO:     true,
		MaxTotal: 4,
		MaxIdle:  4,
		MinIdle:  0,
		//SoftMinEvictableIdleTime: 60 * time.Second,
		//TimeBetweenEvictionRuns:  10 * time.Second,
	})
	dp.pluginPool.StartEvictor()

	return dp
}

func (d *Deployment) Close(ctx context.Context) {
	d.pluginPool.Close(ctx)
}

func (d *Deployment) pluginFactory(ctx context.Context) (interface{}, error) {
	p, err := module.New(ctx, d.wasm, d.config, d.env)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (d *Deployment) BorrowPlugin(ctx context.Context) (*module.Module, error) {
	obj, err := d.pluginPool.BorrowObject(ctx)
	if err != nil {
		return nil, err
	}
	return obj.(*module.Module), nil
}

func (d *Deployment) ReturnPlugin(ctx context.Context, p *module.Module) error {
	return d.pluginPool.ReturnObject(ctx, p)
}

func (d *Deployment) InvalidatePlugin(ctx context.Context, p *module.Module) error {
	return d.pluginPool.InvalidateObject(ctx, p)
}

func (d *Deployment) HandleEvent(ctx context.Context, event *event.Event) error {
	if !slices.Contains(d.Manifest().Events, event.Type) {
		return nil
	}

	plugin, err := d.BorrowPlugin(ctx)
	if err != nil {
		return fmt.Errorf("failed to borrow plugin: %w", err)
	}

	res, err := plugin.Handle(ctx, event)

	fmt.Println("Execution duration: ", res.ExecutionDuration)

	d.env.TrackEventHandled(ctx, string(event.Type), err == nil, res.TotalDuration, res.ExecutionDuration)

	if err != nil {
		slog.With(logattr.Error(err)).Error("failed to handle event")

		// TODO: think about other error types that should invalidate the plugin (e.g. panic / wasm trap)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			if err := d.InvalidatePlugin(ctx, plugin); err != nil {
				slog.With(logattr.Error(err)).Error("failed to invalidate plugin")
			}
		} else {
			if err := d.ReturnPlugin(ctx, plugin); err != nil {
				slog.With(logattr.Error(err)).Error("failed to return plugin")
			}
		}

		d.env.Log(ctx, log.LogLevelError, err.Error())
	} else {
		if err := d.ReturnPlugin(ctx, plugin); err != nil {
			slog.With(logattr.Error(err)).Error("failed to return plugin")
		}
	}

	return nil
}
