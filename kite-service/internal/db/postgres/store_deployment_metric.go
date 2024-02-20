package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
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
		Metadata:           rawMetadata.RawMessage,
		EventType:          entry.EventType,
		EventSuccess:       entry.EventSuccess,
		EventExecutionTime: entry.EventExecutionTime.Microseconds(),
		EventTotalTime:     entry.EventTotalTime.Microseconds(),
		CallType:           entry.CallType,
		CallSuccess:        entry.CallSuccess,
		CallTotalTime:      entry.CallTotalTime.Microseconds(),
		Timestamp:          timeToTimestamp(entry.Timestamp),
	})
	return err
}

func (c *Client) GetDeploymentsMetricsSummary(ctx context.Context, guildID string, startAt time.Time, endAt time.Time) (model.DeploymentMetricsSummary, error) {
	row, err := c.Q.GetDeploymentsMetricsSummary(ctx, pgmodel.GetDeploymentsMetricsSummaryParams{
		GuildID: guildID,
		StartAt: timeToTimestamp(startAt),
		EndAt:   timeToTimestamp(endAt),
	})
	if err != nil {
		return model.DeploymentMetricsSummary{}, err
	}

	return model.DeploymentMetricsSummary{
		TotalCount:                int(row.TotalCount),
		SuccessCount:              int(row.SuccessCount),
		AverageEventExecutionTime: time.Duration(row.AvgEventExecutionTime) * time.Microsecond,
		TotalEventExecutionTime:   time.Duration(row.TotalEventExecutionTime) * time.Microsecond,
		AverageEventTotalTime:     time.Duration(row.AvgEventTotalTime) * time.Microsecond,
		TotalEventTotalTime:       time.Duration(row.TotalEventTotalTime) * time.Microsecond,
		AverageCallTotalTime:      time.Duration(row.AvgCallTotalTime) * time.Microsecond,
	}, nil

}

func (c *Client) GetDeploymentEventMetrics(ctx context.Context, deploymentID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentEventMetricEntry, error) {
	precision, step, err := groupByToPrecisionAndStep(groupBy)
	if err != nil {
		return nil, err
	}

	rows, err := c.Q.GetDeploymentEventMetrics(ctx, pgmodel.GetDeploymentEventMetricsParams{
		DeploymentID: deploymentID,
		StartAt:      timeToTimestamp(startAt),
		EndAt:        timeToTimestamp(time.Now().UTC()),
		Precision:    precision,
		SeriesStep:   step,
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.DeploymentEventMetricEntry, len(rows))
	for i, row := range rows {
		res[i] = model.DeploymentEventMetricEntry{
			Timestamp:            row.Timestamp.Time,
			TotalCount:           int(row.TotalCount),
			SuccessCount:         int(row.SuccessCount),
			AverageExecutionTime: time.Duration(row.AvgExecutionTime) * time.Microsecond,
			AverageTotalTime:     time.Duration(row.AvgTotalTime) * time.Microsecond,
		}
	}

	return res, nil
}

func (c *Client) GetDeploymentCallMetrics(ctx context.Context, deploymentID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentCallMetricEntry, error) {
	precision, step, err := groupByToPrecisionAndStep(groupBy)
	if err != nil {
		return nil, err
	}

	rows, err := c.Q.GetDeploymentCallMetrics(ctx, pgmodel.GetDeploymentCallMetricsParams{
		DeploymentID: deploymentID,
		StartAt:      timeToTimestamp(startAt),
		EndAt:        timeToTimestamp(time.Now().UTC()),
		Precision:    precision,
		SeriesStep:   step,
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.DeploymentCallMetricEntry, len(rows))
	for i, row := range rows {
		res[i] = model.DeploymentCallMetricEntry{
			Timestamp:        row.Timestamp.Time,
			TotalCount:       int(row.TotalCount),
			SuccessCount:     int(row.SuccessCount),
			AverageTotalTime: time.Duration(row.AvgTotalTime) * time.Microsecond,
		}
	}

	return res, nil
}

func (c *Client) GetDeploymentsEventMetrics(ctx context.Context, guildID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentEventMetricEntry, error) {
	precision, step, err := groupByToPrecisionAndStep(groupBy)
	if err != nil {
		return nil, err
	}

	rows, err := c.Q.GetDeploymentsEventMetrics(ctx, pgmodel.GetDeploymentsEventMetricsParams{
		GuildID:    guildID,
		StartAt:    timeToTimestamp(startAt),
		EndAt:      timeToTimestamp(time.Now().UTC()),
		Precision:  precision,
		SeriesStep: step,
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.DeploymentEventMetricEntry, len(rows))
	for i, row := range rows {
		res[i] = model.DeploymentEventMetricEntry{
			Timestamp:            row.Timestamp.Time,
			TotalCount:           int(row.TotalCount),
			SuccessCount:         int(row.SuccessCount),
			AverageExecutionTime: time.Duration(row.AvgExecutionTime) * time.Microsecond,
			AverageTotalTime:     time.Duration(row.AvgTotalTime) * time.Microsecond,
		}
	}

	return res, nil
}

func (c *Client) GetDeploymentsCallMetrics(ctx context.Context, guildID string, startAt time.Time, groupBy time.Duration) ([]model.DeploymentCallMetricEntry, error) {
	precision, step, err := groupByToPrecisionAndStep(groupBy)
	if err != nil {
		return nil, err
	}

	rows, err := c.Q.GetDeploymentsCallMetrics(ctx, pgmodel.GetDeploymentsCallMetricsParams{
		GuildID:    guildID,
		StartAt:    timeToTimestamp(startAt),
		EndAt:      timeToTimestamp(time.Now().UTC()),
		Precision:  precision,
		SeriesStep: step,
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.DeploymentCallMetricEntry, len(rows))
	for i, row := range rows {
		res[i] = model.DeploymentCallMetricEntry{
			Timestamp:        row.Timestamp.Time,
			TotalCount:       int(row.TotalCount),
			SuccessCount:     int(row.SuccessCount),
			AverageTotalTime: time.Duration(row.AvgTotalTime) * time.Microsecond,
		}
	}

	return res, nil
}

func groupByToPrecisionAndStep(groupBy time.Duration) (string, string, error) {
	switch groupBy {
	case time.Hour * 24 * 7:
		return "week", "7d", nil
	case time.Hour * 24:
		return "day", "1d", nil
	case time.Hour:
		return "hour", "1h", nil
	case time.Minute:
		return "minute", "1m", nil
	case time.Second:
		return "second", "1s", nil
	default:
		return "", "", fmt.Errorf("unsupported group by duration: %s", groupBy)
	}
}
