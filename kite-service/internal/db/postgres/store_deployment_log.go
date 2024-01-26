package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

func (c *Client) GetDeploymentLogEntries(ctx context.Context, deploymentID string, guildID string) ([]model.DeploymentLogEntry, error) {
	// TODO: use guild id to validate that deployment belongs to that guild

	entries, err := c.Q.GetDeploymentLogEntries(ctx, deploymentID)
	if err != nil {
		return nil, err
	}

	res := make([]model.DeploymentLogEntry, len(entries))
	for i, entry := range entries {
		res[i] = model.DeploymentLogEntry{
			ID:           uint64(entry.ID),
			DeploymentID: entry.DeploymentID,
			Level:        entry.Level,
			Message:      entry.Message,
			CreatedAt:    entry.CreatedAt,
		}
	}

	return res, nil
}

func (c *Client) GetDeploymentLogSummary(ctx context.Context, deploymentID string, guildID string, cutoff time.Time) (*model.DeploymentLogSummary, error) {
	summary, err := c.Q.GetDeploymentLogSummary(ctx, pgmodel.GetDeploymentLogSummaryParams{
		DeploymentID: deploymentID,
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
