package bundlereview

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// HandleOpenBundleReportList lists the reports waiting for a decision, oldest
// first, with who made the bundle and who reported it.
func (h *BundleReviewHandler) HandleOpenBundleReportList(c *handler.Context) (*wire.OpenBundleReportListResponse, error) {
	reports, err := h.bundleStore.OpenBundleReports(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle reports: %w", err)
	}

	users := map[string]*wire.BundleOwner{}
	user := func(id string) (*wire.BundleOwner, error) {
		if u, ok := users[id]; ok {
			return u, nil
		}
		u, err := h.userStore.User(c.Context(), id)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		users[id] = wire.BundleOwnerToWire(u)
		return users[id], nil
	}

	res := make([]*wire.OpenBundleReport, len(reports))
	for i, r := range reports {
		owner, err := user(r.Bundle.OwnerUserID.String)
		if err != nil {
			return nil, err
		}
		reporter, err := user(r.Report.ReporterUserID)
		if err != nil {
			return nil, err
		}
		res[i] = &wire.OpenBundleReport{
			ID:                r.Report.ID,
			BundleID:          r.Bundle.ID,
			BundleName:        r.Bundle.Name,
			BundleDescription: r.Bundle.Description,
			Owner:             owner,
			Reporter:          reporter,
			Category:          string(r.Report.Category),
			Details:           r.Report.Details,
			CreatedAt:         r.Report.CreatedAt,
		}
	}
	return &res, nil
}

// HandleBundleReportDismiss resolves a report without taking its bundle down.
func (h *BundleReviewHandler) HandleBundleReportDismiss(c *handler.Context) (*wire.BundleReportDismissResponse, error) {
	err := h.bundleStore.DismissBundleReport(c.Context(), c.Param("reportID"), c.Session.UserID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle_report", "the report doesn't exist or was resolved already")
		}
		return nil, fmt.Errorf("failed to dismiss bundle report: %w", err)
	}
	return &wire.BundleReportDismissResponse{}, nil
}

// HandleBundleRemove takes a user's bundle down for breaking the rules. Only
// its owner still sees it, with the reason, and apps that added it keep what
// it added without updates.
func (h *BundleReviewHandler) HandleBundleRemove(c *handler.Context, req wire.BundleRemoveRequest) (*wire.BundleRemoveResponse, error) {
	err := h.bundleStore.RemoveBundle(c.Context(), c.Param("bundleID"), req.Reason, c.Session.UserID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle", "the template doesn't exist, is official or was taken down already")
		}
		return nil, fmt.Errorf("failed to remove bundle: %w", err)
	}
	return &wire.BundleRemoveResponse{}, nil
}
