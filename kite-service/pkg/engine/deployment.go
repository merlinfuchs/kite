package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-sdk-go/log"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/module"
	"github.com/tetratelabs/wazero/sys"
)

type Deployment struct {
	ID       string
	wasm     []byte
	manifest manifest.Manifest
	config   DeploymentConfig
	env      module.HostEnvironment

	modulePool *pool.ObjectPool
}

type DeploymentConfig struct {
	module.ModuleConfig
	PoolMaxTotal int
	PoolMaxIdle  int
	PoolMinIdle  int
}

func (d *Deployment) Manifest() manifest.Manifest {
	return d.manifest
}

func (d *Deployment) Config() DeploymentConfig {
	return d.config
}

func (d *Deployment) ModuleConfig() module.ModuleConfig {
	return d.config.ModuleConfig
}

func NewDeployment(
	id string,
	env module.HostEnvironment,
	wasm []byte,
	manifest manifest.Manifest,
	config DeploymentConfig,
) *Deployment {
	dp := &Deployment{
		ID:       id,
		env:      env,
		wasm:     wasm,
		manifest: manifest,
		config:   config,
	}

	factory := pool.NewPooledObjectFactorySimple(dp.moduleFactory)
	dp.modulePool = pool.NewObjectPool(context.Background(), factory, &pool.ObjectPoolConfig{
		LIFO:                     true,
		MaxTotal:                 config.PoolMaxTotal,
		MaxIdle:                  config.PoolMaxIdle,
		MinIdle:                  config.PoolMinIdle,
		SoftMinEvictableIdleTime: 60 * time.Second,
		TimeBetweenEvictionRuns:  10 * time.Second,
	})
	// TODO? This spawns a goroutine for each deployment, maybe we should have a single evictor for all deployments
	dp.modulePool.StartEvictor()

	return dp
}

func (d *Deployment) Close(ctx context.Context) {
	d.modulePool.Close(ctx)
}

func (d *Deployment) moduleFactory(ctx context.Context) (interface{}, error) {
	p, err := module.New(ctx, d.wasm, d.config.ModuleConfig, d.env)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (d *Deployment) BorrowModule(ctx context.Context) (*module.Module, error) {
	obj, err := d.modulePool.BorrowObject(ctx)
	if err != nil {
		return nil, err
	}
	return obj.(*module.Module), nil
}

func (d *Deployment) ReturnModule(ctx context.Context, p *module.Module) error {
	return d.modulePool.ReturnObject(ctx, p)
}

func (d *Deployment) InvalidateModule(ctx context.Context, p *module.Module) error {
	return d.modulePool.InvalidateObject(ctx, p)
}

func (d *Deployment) HandleEvent(ctx context.Context, event *event.Event) error {
	if !slices.Contains(d.Manifest().Events, event.Type) {
		return nil
	}

	module, err := d.BorrowModule(ctx)
	if err != nil {
		return fmt.Errorf("failed to borrow module: %w", err)
	}

	res, err := module.Handle(ctx, event)

	fmt.Println("Execution duration: ", res.ExecutionDuration)

	d.env.TrackEventHandled(ctx, string(event.Type), err == nil, res.TotalDuration, res.ExecutionDuration)

	if err != nil {
		slog.With(logattr.Error(err)).Error("failed to handle event")

		// TODO: think about other error types that should invalidate the module (e.g. panic / wasm trap) (or just all errors?)
		_, isExitError := err.(*sys.ExitError)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || isExitError {
			if err := d.InvalidateModule(ctx, module); err != nil {
				slog.With(logattr.Error(err)).Error("failed to invalidate module")
			}
		} else {
			if err := d.ReturnModule(ctx, module); err != nil {
				slog.With(logattr.Error(err)).Error("failed to return module")
			}
		}

		d.env.Log(ctx, log.LogLevelError, err.Error())
	} else {
		if err := d.ReturnModule(ctx, module); err != nil {
			slog.With(logattr.Error(err)).Error("failed to return module")
		}
	}

	return nil
}
