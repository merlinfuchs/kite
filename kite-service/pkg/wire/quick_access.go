package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type QuickAccessItem struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QuickAccessItemListResponse APIResponse[[]QuickAccessItem]

func QuickAccessItemToWire(item *model.QuickAccessItem) QuickAccessItem {
	return QuickAccessItem{
		ID:        item.ID,
		AppID:     item.AppID,
		Type:      string(item.Type),
		Name:      item.Name,
		UpdatedAt: item.UpdatedAt,
	}
}
