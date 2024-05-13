package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.DeploymentStore = (*Client)(nil)

func (c *Client) UpsertDeployment(ctx context.Context, deployment model.Deployment) (*model.Deployment, error) {
	rawManifest, err := json.Marshal(deployment.Manifest)
	if err != nil {
		return nil, err
	}

	rawConfig, err := json.Marshal(deployment.Config)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.UpsertDeployment(ctx, pgmodel.UpsertDeploymentParams{
		ID:              deployment.ID,
		Name:            deployment.Name,
		Key:             deployment.Key,
		Description:     deployment.Description,
		AppID:           deployment.AppID,
		PluginVersionID: nullStringToText(deployment.PluginVersionID),
		WasmBytes:       deployment.WasmBytes,
		Manifest:        rawManifest,
		Config:          rawConfig,
		CreatedAt:       timeToTimestamp(deployment.CreatedAt),
		UpdatedAt:       timeToTimestamp(deployment.UpdatedAt),
	})
	if err != nil {
		return nil, err
	}

	res, err := deploymentToModel(row)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) DeleteDeployment(ctx context.Context, id string, appID string) error {
	_, err := c.Q.DeleteDeployment(ctx, pgmodel.DeleteDeploymentParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) GetDeployment(ctx context.Context, id string, appID string) (*model.Deployment, error) {
	row, err := c.Q.GetDeploymentForApp(ctx, pgmodel.GetDeploymentForAppParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	res, err := deploymentToModel(row)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) GetDeployments(ctx context.Context) ([]model.Deployment, error) {
	rows, err := c.Q.GetDeployments(ctx)
	if err != nil {
		return nil, err
	}

	deployments := make([]model.Deployment, len(rows))
	for i, row := range rows {
		deployments[i], err = deploymentToModel(row)
		if err != nil {
			return nil, err
		}
	}

	return deployments, nil
}

func (c *Client) GetDeploymentsForApp(ctx context.Context, appID string) ([]model.Deployment, error) {
	rows, err := c.Q.GetDeploymentsForApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	deployments := make([]model.Deployment, len(rows))
	for i, row := range rows {
		deployments[i], err = deploymentToModel(row)
		if err != nil {
			return nil, err
		}
	}

	return deployments, nil
}

func (c *Client) GetAppIDsWithDeployment(ctx context.Context) ([]string, error) {
	return c.Q.GetAppIdsWithDeployments(ctx)
}

func (c *Client) GetDeploymentsWithUndeployedChanges(ctx context.Context) ([]model.Deployment, error) {
	rows, err := c.Q.GetDeploymentsWithUndeployedChanges(ctx)
	if err != nil {
		return nil, err
	}

	deployments := make([]model.Deployment, len(rows))
	for i, row := range rows {
		deployments[i], err = deploymentToModel(row)
		if err != nil {
			return nil, err
		}
	}

	return deployments, nil
}

func (c *Client) GetDeploymentIDs(ctx context.Context) ([]model.PartialDeployment, error) {
	rows, err := c.Q.GetDeploymentIDs(ctx)
	if err != nil {
		return nil, err
	}

	deployments := make([]model.PartialDeployment, len(rows))
	for i, row := range rows {
		deployments[i] = model.PartialDeployment{
			ID:    row.ID,
			AppID: row.AppID,
		}
	}

	return deployments, nil
}

func (c *Client) UpdateDeploymentsDeployedAtForApp(ctx context.Context, appID string, deployedAt time.Time) error {
	_, err := c.Q.UpdateDeploymentsDeployedAtForApp(ctx, pgmodel.UpdateDeploymentsDeployedAtForAppParams{
		AppID:      appID,
		DeployedAt: timeToTimestamp(deployedAt),
	})
	if err != nil {
		return err
	}

	return nil
}

func deploymentToModel(deployment pgmodel.Deployment) (model.Deployment, error) {
	manifest := manifest.Manifest{}
	err := json.Unmarshal(deployment.Manifest, &manifest)
	if err != nil {
		return model.Deployment{}, err
	}

	config := make(map[string]string)
	err = json.Unmarshal(deployment.Config, &config)
	if err != nil {
		return model.Deployment{}, err
	}

	return model.Deployment{
		ID:              deployment.ID,
		Name:            deployment.Name,
		Key:             deployment.Key,
		Description:     deployment.Description,
		AppID:           deployment.AppID,
		PluginVersionID: textToNullString(deployment.PluginVersionID),
		WasmBytes:       deployment.WasmBytes,
		Manifest:        manifest,
		Config:          config,
		CreatedAt:       deployment.CreatedAt.Time,
		UpdatedAt:       deployment.UpdatedAt.Time,
		DeployedAt:      timestampToNullTime(deployment.DeployedAt),
	}, nil
}
