package postgres

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) UpsertDeployment(ctx context.Context, deployment model.Deployment) (*model.Deployment, error) {
	res, err := c.Q.UpsertDeployment(ctx, pgmodel.UpsertDeploymentParams{
		ID:              deployment.ID,
		Name:            deployment.Name,
		Key:             deployment.Key,
		Description:     deployment.Description,
		GuildID:         deployment.GuildID,
		PluginVersionID: deployment.PluginVersionID.NullString,
		WasmBytes:       deployment.WasmBytes,
		//ManifestDefaultConfig: deployment.ManifestDefaultConfig,
		ManifestEvents:   deployment.ManifestEvents,
		ManifestCommands: deployment.ManifestCommands,
		//Config:                deployment.Config,
		CreatedAt: deployment.CreatedAt,
		UpdatedAt: deployment.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}

	return deploymentToModel(res), nil
}

func (c *Client) DeleteDeployment(ctx context.Context, deployment model.Deployment) error {
	return nil
}

func (c *Client) GetDeploymentForGuild(ctx context.Context, id string, guildID string) (*model.Deployment, error) {
	return nil, nil
}

func (c *Client) GetDeploymentsForGuild(ctx context.Context, guildID string) ([]model.Deployment, error) {
	rows, err := c.Q.GetDeploymentsForGuild(ctx, guildID)
	if err != nil {
		return nil, err
	}

	deployments := make([]model.Deployment, len(rows))
	for i, row := range rows {
		deployments[i] = *deploymentToModel(row)
	}

	return deployments, nil
}

func (c *Client) GetGuildIDsWithDeployment(ctx context.Context) ([]string, error) {
	return c.Q.GetGuildIdsWithDeployments(ctx)
}

func deploymentToModel(deployment pgmodel.Deployment) *model.Deployment {
	return &model.Deployment{
		ID:          deployment.ID,
		Name:        deployment.Name,
		Key:         deployment.Key,
		Description: deployment.Description,
		GuildID:     deployment.GuildID,
		PluginVersionID: null.String{
			NullString: deployment.PluginVersionID,
		},
		WasmBytes: deployment.WasmBytes,
		//ManifestDefaultConfig: deployment.ManifestDefaultConfig,
		ManifestEvents:   deployment.ManifestEvents,
		ManifestCommands: deployment.ManifestCommands,
		//Config:                deployment.Config,
		CreatedAt: deployment.CreatedAt,
		UpdatedAt: deployment.UpdatedAt,
	}
}
