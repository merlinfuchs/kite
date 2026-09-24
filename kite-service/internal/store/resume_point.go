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
}
