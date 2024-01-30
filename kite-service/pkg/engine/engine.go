package engine

import (
	"context"
	"slices"
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

func (e *PluginEngine) RemoveDeployment(ctx context.Context, guildID string, deploymentID string) {
	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[guildID]
	if !exists {
		return
	}

	for i, deployment := range deployments {
		if deployment.ID == deploymentID {
			deployment.Close(ctx)
			deployments = append(deployments[:i], deployments[i+1:]...)
			e.Deployments[guildID] = deployments
			return
		}
	}
}

func (e *PluginEngine) LoadGuildDeployment(ctx context.Context, guildID string, deployment *PluginDeployment) {
	e.RemoveDeployment(ctx, guildID, deployment.ID)

	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[guildID]
	if !exists {
		deployments = []*PluginDeployment{}
	}

	deployments = append(deployments, deployment)
	e.Deployments[guildID] = deployments
}

func (e *PluginEngine) ReplaceGuildDeployments(ctx context.Context, guildID string, deployments []*PluginDeployment) {
	e.Lock()
	defer e.Unlock()

	existing := e.Deployments[guildID]
	for _, deployment := range existing {
		deployment.Close(ctx)
	}

	e.Deployments[guildID] = deployments
}

func (e *PluginEngine) TruncateGuildDeployments(ctx context.Context, guildID string, deploymentIDs []string) bool {
	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[guildID]
	if !exists {
		return false
	}

	removed := false

	for i, deployment := range deployments {
		if !slices.Contains(deploymentIDs, deployment.ID) {
			deployment.Close(ctx)
			deployments = append(deployments[:i], deployments[i+1:]...)
			removed = true
		}
	}

	e.Deployments[guildID] = deployments
	return removed
}
