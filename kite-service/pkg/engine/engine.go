package engine

import (
	"context"
	"slices"
	"sync"

	"github.com/merlinfuchs/dismod/distype"
)

type Engine struct {
	sync.RWMutex

	Deployments map[distype.Snowflake][]*Deployment
}

func New() *Engine {
	return &Engine{
		Deployments: map[distype.Snowflake][]*Deployment{},
	}
}

func (e *Engine) RemoveDeployment(ctx context.Context, guildID distype.Snowflake, deploymentID string) {
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

func (e *Engine) LoadGuildDeployment(ctx context.Context, guildID distype.Snowflake, deployment *Deployment) {
	e.RemoveDeployment(ctx, guildID, deployment.ID)

	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[guildID]
	if !exists {
		deployments = []*Deployment{}
	}

	deployments = append(deployments, deployment)
	e.Deployments[guildID] = deployments
}

func (e *Engine) ReplaceGuildDeployments(ctx context.Context, guildID distype.Snowflake, deployments []*Deployment) {
	e.Lock()
	defer e.Unlock()

	existing := e.Deployments[guildID]
	for _, deployment := range existing {
		deployment.Close(ctx)
	}

	e.Deployments[guildID] = deployments
}

func (e *Engine) TruncateGuildDeployments(ctx context.Context, guildID distype.Snowflake, deploymentIDs []string) bool {
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
