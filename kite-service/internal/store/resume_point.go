package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ResumePointStore interface {
	CreateResumePoint(ctx context.Context, resumePoint *model.ResumePoint) error
	DeleteResumePoint(ctx context.Context, id string) error
	// DeleteStaleResumePoints deletes up to batchSize resume points that expired or
	// were last used before usedBefore and returns how many were deleted.
	DeleteStaleResumePoints(ctx context.Context, now time.Time, usedBefore time.Time, batchSize int) (int64, error)
	TouchResumePoint(ctx context.Context, appID string, id string, usedAt time.Time) error
	ResumePoint(ctx context.Context, appID string, id string) (*model.ResumePoint, error)
	CountPendingTimerResumePoints(ctx context.Context, appID string) (int, error)
	HasDueTimerResumePoints(ctx context.Context, now time.Time) (bool, error)
	// LeaseDueTimerResumePoints returns up to batchSize timers of the given
	// apps that are due at now, and moves them to leaseUntil so no one else
	// resumes them in the meantime.
	LeaseDueTimerResumePoints(ctx context.Context, appIDs []string, now time.Time, leaseUntil time.Time, batchSize int) ([]*model.ResumePoint, error)
	// DeleteTimerResumePoint reports whether it deleted the timer. Only the
	// caller that deleted it may resume the flow.
	DeleteTimerResumePoint(ctx context.Context, appID string, id string) (bool, error)
}
