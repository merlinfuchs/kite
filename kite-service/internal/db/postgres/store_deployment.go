package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/merlinfuchs/kite/go-types/manifest"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"gopkg.in/guregu/null.v4"
)

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
		GuildID:         deployment.GuildID,
		PluginVersionID: deployment.PluginVersionID.NullString,
		WasmBytes:       deployment.WasmBytes,
		Manifest:        rawManifest,
		Config:          rawConfig,
		CreatedAt:       deployment.CreatedAt,
		UpdatedAt:       deployment.UpdatedAt,
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

func (c *Client) DeleteDeployment(ctx context.Context, id string, guildID string) error {
	_, err := c.Q.DeleteDeployment(ctx, pgmodel.DeleteDeploymentParams{
		ID:      id,
		GuildID: guildID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) GetDeployment(ctx context.Context, id string, guildID string) (*model.Deployment, error) {
	row, err := c.Q.GetDeploymentForGuild(ctx, pgmodel.GetDeploymentForGuildParams{
		ID:      id,
		GuildID: guildID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
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

func (c *Client) GetDeploymentsForGuild(ctx context.Context, guildID string) ([]model.Deployment, error) {
	rows, err := c.Q.GetDeploymentsForGuild(ctx, guildID)
	if err != nil {
		return nil, err
	}

	deployments := make([]model.Deployment, len(rows))
	for i, row := range rows {
		deployments[i], err = deploymentToModel(row)
	}

	return deployments, nil
}

func (c *Client) GetGuildIDsWithDeployment(ctx context.Context) ([]string, error) {
	return c.Q.GetGuildIdsWithDeployments(ctx)
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
		ID:          deployment.ID,
		Name:        deployment.Name,
		Key:         deployment.Key,
		Description: deployment.Description,
		GuildID:     deployment.GuildID,
		PluginVersionID: null.String{
			NullString: deployment.PluginVersionID,
		},
		WasmBytes: deployment.WasmBytes,
		Manifest:  manifest,
		Config:    config,
		CreatedAt: deployment.CreatedAt,
		UpdatedAt: deployment.UpdatedAt,
	}, nil
}
