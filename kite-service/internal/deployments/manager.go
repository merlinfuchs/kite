package deployments

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type DeploymentManager struct {
	store  store.DeploymentStore
	engine *engine.PluginEngine

	stopped chan struct{}
}

func NewManager(store store.DeploymentStore, engine *engine.PluginEngine) *DeploymentManager {
	return &DeploymentManager{
		store:  store,
		engine: engine,
	}
}

func (m *DeploymentManager) Start() {
	m.stopped = make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Minute)
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
