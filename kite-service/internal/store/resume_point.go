package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ResumePointStore interface {
	CreateResumePoint(ctx context.Context, resumePoint *model.ResumePoint) (*model.ResumePoint, error)
	DeleteResumePoint(ctx context.Context, id string) error
	DeleteExpiredResumePoints(ctx context.Context, timestamp time.Time) error
	ResumePoint(ctx context.Context, id string) (*model.ResumePoint, error)
}
