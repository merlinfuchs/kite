package postgres

import (
	"context"

	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.QuickAccessStore = (*Client)(nil)

func (c *Client) GetQuickAccessItems(ctx context.Context, appID string, limit int) ([]model.QuickAccessItem, error) {
	items, err := c.Q.GetQuickAccessItems(ctx, pgmodel.GetQuickAccessItemsParams{
		AppID: appID,
		Limit: int32(limit),
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.QuickAccessItem, len(items))
	for i, item := range items {
		res[i] = model.QuickAccessItem{
			ID:        item.ID,
			AppID:     item.AppID,
			Type:      model.QuickAccessItemType(item.Type),
			Name:      item.Name,
			UpdatedAt: item.UpdatedAt.Time,
		}
	}

	return res, nil
}
