package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ResumePointStore interface {
	CreateResumePoint(ctx context.Context, resumePoint *model.ResumePoint) error
	DeleteResumePoint(ctx context.Context, id string) error
	DeleteExpiredResumePoints(ctx context.Context, timestamp time.Time) error
	// ResumePoint is scoped by app: the ID reaches us from an interaction's
	// custom_id, so a resume point belonging to another app has to be
	// indistinguishable from one that does not exist.
	ResumePoint(ctx context.Context, appID string, id string) (*model.ResumePoint, error)
}
