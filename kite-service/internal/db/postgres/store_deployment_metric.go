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
		EventID:            int64(entry.EventID),
		EventType:          entry.EventType,
		EventSuccess:       entry.EventSuccess,
		EventExecutionTime: int32(entry.EventExecutionTime.Milliseconds()),
		EventTotalTime:     int32(entry.EventTotalTime.Milliseconds()),
		CallType:           entry.CallType,
		CallSuccess:        entry.CallSuccess,
		CallTotalTime:      int32(entry.CallTotalTime.Milliseconds()),
		Timestamp:          entry.Timestamp,
	})
	return err
}

func (c *Client) GetDeploymentMetricEntries(ctx context.Context, args store.GetDeploymentMetricEntriesArgs) ([]model.DeploymentMetricEntry, error) {
	return nil, nil
}
