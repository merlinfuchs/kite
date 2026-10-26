package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) SubmitBundleReport(ctx context.Context, report *model.BundleReport) error {
	return c.Q.UpsertOpenBundleReport(ctx, pgmodel.UpsertOpenBundleReportParams{
		ID:             report.ID,
		BundleID:       report.BundleID,
		ReporterUserID: report.ReporterUserID,
		Category:       string(report.Category),
		Details:        report.Details,
		CreatedAt:      pgtype.Timestamp{Time: report.CreatedAt.UTC(), Valid: true},
	})
}

func (c *Client) OpenBundleReports(ctx context.Context) ([]*model.OpenBundleReport, error) {
	rows, err := c.Q.GetOpenBundleReports(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*model.OpenBundleReport, len(rows))
	for i, row := range rows {
		b, err := rowToBundle(row.Bundle, nil, nil)
		if err != nil {
			return nil, err
		}
		res[i] = &model.OpenBundleReport{Report: rowToBundleReport(row.BundleReport), Bundle: b}
	}
	return res, nil
}

func (c *Client) DismissBundleReport(ctx context.Context, id string, resolverUserID string, at time.Time) error {
	rows, err := c.Q.ResolveBundleReport(ctx, pgmodel.ResolveBundleReportParams{
		ID:             id,
		ResolvedAt:     pgtype.Timestamp{Time: at.UTC(), Valid: true},
		ResolverUserID: pgtype.Text{String: resolverUserID, Valid: true},
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (c *Client) RemoveBundle(ctx context.Context, id string, reason string, resolverUserID string, at time.Time) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	rows, err := q.RemoveBundle(ctx, pgmodel.RemoveBundleParams{
		ID:            id,
		RemovedAt:     pgtype.Timestamp{Time: at.UTC(), Valid: true},
		RemovedReason: reason,
	})
	if err != nil {
		return fmt.Errorf("failed to remove bundle: %w", err)
	}
	if rows == 0 {
		return store.ErrNotFound
	}

	// Its reports are dealt with, and it leaves the marketplace for good.
	err = q.ResolveBundleReportsOfBundle(ctx, pgmodel.ResolveBundleReportsOfBundleParams{
		BundleID:       id,
		ResolvedAt:     pgtype.Timestamp{Time: at.UTC(), Valid: true},
		ResolverUserID: pgtype.Text{String: resolverUserID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to resolve bundle reports: %w", err)
	}
	if err := unlistBundle(ctx, q, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func rowToBundleReport(row pgmodel.BundleReport) *model.BundleReport {
	return &model.BundleReport{
		ID:             row.ID,
		BundleID:       row.BundleID,
		ReporterUserID: row.ReporterUserID,
		Category:       model.BundleReportCategory(row.Category),
		Details:        row.Details,
		CreatedAt:      row.CreatedAt.Time,
		ResolvedAt:     null.NewTime(row.ResolvedAt.Time, row.ResolvedAt.Valid),
		ResolverUserID: null.NewString(row.ResolverUserID.String, row.ResolverUserID.Valid),
	}
}
