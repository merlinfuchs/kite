package deployments

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/tetratelabs/wazero"
)

type DeploymentManager struct {
	store            store.DeploymentStore
	engine           *engine.Engine
	envStores        host.HostEnvironmentStores
	apps             store.AppProvider
	compilationCache wazero.CompilationCache
	limits           config.ServerEngineLimitConfig
}

func NewManager(
	store store.DeploymentStore,
	engine *engine.Engine,
	envStores host.HostEnvironmentStores,
	apps store.AppProvider,
	limits config.ServerEngineLimitConfig,
) (*DeploymentManager, error) {
	compilationCache, err := wazero.NewCompilationCacheWithDir("./.wasm-compilation-cache")
	if err != nil {
		return nil, fmt.Errorf("error creating wazero compilation cache: %w", err)
	}

	return &DeploymentManager{
		store:            store,
		engine:           engine,
		envStores:        envStores,
		compilationCache: compilationCache,
		apps:             apps,
		limits:           limits,
	}, nil
}

func (m *DeploymentManager) Start(ctx context.Context) error {
	err := m.populateEngineDeployments(ctx)
	if err != nil {
		return fmt.Errorf("error populating engine deployments: %w", err)
	}

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := m.updateEngineDeployments(ctx)
				if err != nil {
					slog.With(logattr.Error(err)).Error("Error updating engine deployments")
				}
			}
		}
	}()

	return nil
}
