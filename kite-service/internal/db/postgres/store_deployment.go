package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
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

func (c *Client) GetDeploymentLogEntries(ctx context.Context, id string, guildID string) ([]model.DeploymentLogEntry, error) {
	// TODO: use guild id to validate that deployment belongs to that guild

	entries, err := c.Q.GetDeploymentLogEntries(ctx, id)
	if err != nil {
		return nil, err
	}

	res := make([]model.DeploymentLogEntry, len(entries))
	for i, entry := range entries {
		res[i] = model.DeploymentLogEntry{
			ID:           entry.ID,
			DeploymentID: entry.DeploymentID,
			Level:        entry.Level,
			Message:      entry.Message,
			CreatedAt:    entry.CreatedAt,
		}
	}

	return res, nil
}

func (c *Client) GetDeploymentLogSummary(ctx context.Context, id string, guildID string, cutoff time.Time) (*model.DeploymentLogSummary, error) {
	summary, err := c.Q.GetDeploymentLogSummary(ctx, pgmodel.GetDeploymentLogSummaryParams{
		DeploymentID: id,
		CreatedAt:    cutoff,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &model.DeploymentLogSummary{}, nil
		}

		return nil, err
	}

	return &model.DeploymentLogSummary{
		DeploymentID: summary.DeploymentID,
		TotalCount:   int(summary.TotalCount),
		ErrorCount:   int(summary.ErrorCount),
		WarnCount:    int(summary.WarnCount),
		InfoCount:    int(summary.InfoCount),
		DebugCount:   int(summary.DebugCount),
	}, nil
}

func (c *Client) CreateDeploymentLogEntry(ctx context.Context, entry model.DeploymentLogEntry) error {
	err := c.Q.CreateDeploymentLogEntry(ctx, pgmodel.CreateDeploymentLogEntryParams{
		ID:           entry.ID,
		DeploymentID: entry.DeploymentID,
		Level:        entry.Level,
		Message:      entry.Message,
		CreatedAt:    entry.CreatedAt,
	})
	if err != nil {
		return err
	}

	return nil
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
