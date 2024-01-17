package engine

import (
	"sync"

	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
)

type PluginEngine struct {
	sync.RWMutex

	StaticDeployments []*PluginDeployment
	Deployments       map[string][]*PluginDeployment
	env               plugin.HostEnvironment
}

func New(env plugin.HostEnvironment) *PluginEngine {
	return &PluginEngine{
		env:         env,
		Deployments: map[string][]*PluginDeployment{},
	}
}

func (e *PluginEngine) LoadStaticDeployment(deployment *PluginDeployment) error {
	e.Lock()
	defer e.Unlock()

	deployment.env = e.env
	e.StaticDeployments = append(e.StaticDeployments, deployment)
	return nil
}

func (e *PluginEngine) LoadDeployment(guildID string, deployment *PluginDeployment) error {
	e.Lock()
	defer e.Unlock()

	deployment.env = e.env
	deployments, exists := e.Deployments[guildID]
	if !exists {
		deployments = []*PluginDeployment{}
	}

	deployments = append(deployments, deployment)
	e.Deployments[guildID] = deployments
	return nil
}

func (e *PluginEngine) ReplaceGuildDeployments(guildID string, deployments []*PluginDeployment) error {
	e.Lock()
	defer e.Unlock()

	for _, d := range deployments {
		d.env = e.env
	}

	e.Deployments[guildID] = deployments
	return nil
}
