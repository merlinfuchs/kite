package deployments

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/merlinfuchs/dismod/disrest"
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
	discordClient    *disrest.Client
	compilationCache wazero.CompilationCache
	limits           config.ServerEngineLimitConfig

	stopped chan struct{}
}

func NewManager(
	store store.DeploymentStore,
	engine *engine.Engine,
	envStores host.HostEnvironmentStores,
	discordClient *disrest.Client,
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
		discordClient:    discordClient,
		limits:           limits,
	}, nil
}

func (m *DeploymentManager) Start() {
	m.stopped = make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		err := m.populateEngineDeployments(context.Background())
		if err != nil {
			slog.With(logattr.Error(err)).Error("Error populating engine deployments")
			os.Exit(1)
		}

		for {
			select {
			case <-m.stopped:
				return
			case <-ticker.C:
				err := m.updateEngineDeployments(context.Background())
				if err != nil {
					slog.With(logattr.Error(err)).Error("Error updating engine deployments")
				}
			}
		}
	}()
}

func (m *DeploymentManager) Stop() {
	close(m.stopped)
}
