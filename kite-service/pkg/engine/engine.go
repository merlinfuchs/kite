package engine

import (
	"sync"
)

type PluginEngine struct {
	sync.RWMutex

	Deployments map[string][]*PluginDeployment
}

func New() *PluginEngine {
	return &PluginEngine{
		Deployments: map[string][]*PluginDeployment{},
	}
}

func (e *PluginEngine) LoadDeployment(guildID string, deployment *PluginDeployment) error {
	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[guildID]
	if !exists {
		deployments = []*PluginDeployment{}
	}

	deployments = append(deployments, deployment)
	e.Deployments[guildID] = deployments
	return nil
}

func (e *PluginEngine) ReplaceGuildDeployments(guildID string, deployments []*PluginDeployment) {
	e.Lock()
	defer e.Unlock()

	e.Deployments[guildID] = deployments
}
