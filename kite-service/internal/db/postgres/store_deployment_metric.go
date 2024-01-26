package postgres

import (
	"context"
	"encoding/json"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/sqlc-dev/pqtype"
)

func (c *Client) CreateDeploymentMetricEntry(ctx context.Context, entry model.DeploymentMetricEntry) error {
	var rawMetadata pqtype.NullRawMessage
	if entry.Metadata != nil {
		raw, err := json.Marshal(entry.Metadata)
		if err != nil {
			return err
		}
		rawMetadata = pqtype.NullRawMessage{RawMessage: raw, Valid: true}
	}

	err := c.Q.CreateDeploymentMetricEntry(ctx, pgmodel.CreateDeploymentMetricEntryParams{
		DeploymentID:       entry.DeploymentID,
		Type:               string(entry.Type),
		Metadata:           rawMetadata,
		EventType:          entry.EventType,
		EventSuccess:       entry.EventSuccess,
		EventExecutionTime: entry.EventExecutionTime.Microseconds(),
		EventTotalTime:     entry.EventTotalTime.Microseconds(),
		CallType:           entry.CallType,
		CallSuccess:        entry.CallSuccess,
		CallTotalTime:      entry.CallTotalTime.Microseconds(),
		Timestamp:          entry.Timestamp,
	})
	return err
}

func (c *Client) GetDeploymentMetricEntries(ctx context.Context, args store.GetDeploymentMetricEntriesArgs) ([]model.DeploymentMetricEntry, error) {
	return nil, nil
}
