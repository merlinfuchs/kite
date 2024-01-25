package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type QuickAccessStore interface {
	GetQuickAccessItems(ctx context.Context, guildID string, limit int) ([]model.QuickAccessItem, error)
}
