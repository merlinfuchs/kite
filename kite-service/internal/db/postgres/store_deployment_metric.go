package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/sqlc-dev/pqtype"
)

var _ store.DeploymentMetricStore = (*Client)(nil)

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

func (c *Client) GetDeploymentsMetricsSummary(ctx context.Context, appID distype.Snowflake, startAt time.Time, endAt time.Time) (model.DeploymentMetricsSummary, error) {
	row, err := c.Q.GetDeploymentsMetricsSummary(ctx, pgmodel.GetDeploymentsMetricsSummaryParams{
		AppID:   string(appID),
		StartAt: timeToTimestamp(startAt),
		EndAt:   timeToTimestamp(endAt),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.DeploymentMetricsSummary{}, nil
		}
		return model.DeploymentMetricsSummary{}, err
	}

	return model.DeploymentMetricsSummary{
		FirstEntryAt:            row.FirstEntryAt.Time,
		LastEntryAt:             row.LastEntryAt.Time,
		TotalEventCount:         int(row.TotalEventCount),
		SuccessEventCount:       int(row.SuccessEventCount),
		AvgEventExecutionTime:   time.Duration(row.AvgEventExecutionTime) * time.Microsecond,
		TotalEventExecutionTime: time.Duration(row.TotalEventExecutionTime) * time.Microsecond,
		AvgEventTotalTime:       time.Duration(row.AvgEventTotalTime) * time.Microsecond,
		TotalCallCount:          int(row.TotalCallCount),
		SuccessCallCount:        int(row.SuccessCallCount),
		TotalEventTotalTime:     time.Duration(row.TotalEventTotalTime) * time.Microsecond,
		AvgCallTotalTime:        time.Duration(row.AvgCallTotalTime) * time.Microsecond,
		TotalCallTotalTime:      time.Duration(row.TotalCallTotalTime) * time.Microsecond,
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

func (c *Client) GetDeploymentsEventMetrics(ctx context.Context, appID distype.Snowflake, startAt time.Time, groupBy time.Duration) ([]model.DeploymentEventMetricEntry, error) {
	precision, step, err := groupByToPrecisionAndStep(groupBy)
	if err != nil {
		return nil, err
	}

	rows, err := c.Q.GetDeploymentsEventMetrics(ctx, pgmodel.GetDeploymentsEventMetricsParams{
		AppID:      string(appID),
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

func (c *Client) GetDeploymentsCallMetrics(ctx context.Context, appID distype.Snowflake, startAt time.Time, groupBy time.Duration) ([]model.DeploymentCallMetricEntry, error) {
	precision, step, err := groupByToPrecisionAndStep(groupBy)
	if err != nil {
		return nil, err
	}

	rows, err := c.Q.GetDeploymentsCallMetrics(ctx, pgmodel.GetDeploymentsCallMetricsParams{
		AppID:      string(appID),
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
