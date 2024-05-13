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

func (e *Engine) RemoveAppDeployment(ctx context.Context, appID distype.Snowflake, deploymentID string) {
	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[appID]
	if !exists {
		return
	}

	for i, deployment := range deployments {
		if deployment.ID == deploymentID {
			deployment.Close(ctx)
			deployments = append(deployments[:i], deployments[i+1:]...)
			e.Deployments[appID] = deployments
			return
		}
	}
}

func (e *Engine) LoadAppDeployment(ctx context.Context, appID distype.Snowflake, deployment *Deployment) {
	e.RemoveAppDeployment(ctx, appID, deployment.ID)

	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[appID]
	if !exists {
		deployments = []*Deployment{}
	}

	deployments = append(deployments, deployment)
	e.Deployments[appID] = deployments
}

func (e *Engine) ReplaceAppDeployments(ctx context.Context, appID distype.Snowflake, deployments []*Deployment) {
	e.Lock()
	defer e.Unlock()

	existing := e.Deployments[appID]
	for _, deployment := range existing {
		deployment.Close(ctx)
	}

	e.Deployments[appID] = deployments
}

func (e *Engine) TruncateAppDeployments(ctx context.Context, appID distype.Snowflake, deploymentIDs []string) bool {
	e.Lock()
	defer e.Unlock()

	deployments, exists := e.Deployments[appID]
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

	e.Deployments[appID] = deployments
	return removed
}
