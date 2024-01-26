package deployments

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type DeploymentManager struct {
	store     store.DeploymentStore
	engine    *engine.PluginEngine
	envStores host.HostEnvironmentStores

	stopped chan struct{}
}

func NewManager(store store.DeploymentStore, engine *engine.PluginEngine, envStores host.HostEnvironmentStores) *DeploymentManager {
	return &DeploymentManager{
		store:     store,
		engine:    engine,
		envStores: envStores,
	}
}

func (m *DeploymentManager) Start() {
	m.stopped = make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		m.populateEngineDeployments(context.Background())

		for {
			select {
			case <-m.stopped:
				return
			case <-ticker.C:
				m.populateEngineDeployments(context.Background())
			}
		}
	}()
}

func (m *DeploymentManager) Stop() {
	close(m.stopped)
}
