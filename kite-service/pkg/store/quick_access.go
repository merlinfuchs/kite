package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type QuickAccessStore interface {
	GetQuickAccessItems(ctx context.Context, appID string, limit int) ([]model.QuickAccessItem, error)
}
