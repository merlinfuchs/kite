package engine

import (
	"sync"
)

type PluginEngine struct {
	sync.RWMutex

	StaticPlugins []*PluginDeployment
	Deployments   map[string][]*PluginDeployment
}

func New() *PluginEngine {
	return &PluginEngine{}
}

func (e *PluginEngine) LoadStaticPlugin(plugin *PluginDeployment) error {
	e.Lock()
	defer e.Unlock()

	e.StaticPlugins = append(e.StaticPlugins, plugin)
	return nil
}

func (e *PluginEngine) LoadPlugin(p *PluginDeployment, guildID string) error {
	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[guildID]
	if !exists {
		deployments = []*PluginDeployment{}
	}

	deployments = append(deployments, p)
	e.Deployments[guildID] = deployments
	return nil
}
