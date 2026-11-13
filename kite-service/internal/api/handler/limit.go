package handler

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/store"
)

// TotalLimitFactor is how many times its limit an app can have of commands or
// event listeners, counting the disabled ones. Only enabled ones count
// against the limit, so apps can keep some disabled instead of deleting them.
const TotalLimitFactor = 3

// CheckEnabledLimit checks that an app with counts can have added more items,
// addedEnabled of them enabled. name is the plural the error uses. A limit of
// 0 means no limit.
func CheckEnabledLimit(name string, limit int, counts store.ItemCounts, added, addedEnabled int) error {
	if limit == 0 {
		return nil
	}

	// Apps over their limit after a downgrade can still add disabled items.
	if addedEnabled > 0 && counts.Enabled+addedEnabled > limit {
		return ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of enabled %s (%d) reached, disable or delete some first", name, limit))
	}

	if total := limit * TotalLimitFactor; counts.Total+added > total {
		return ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of %s (%d) reached, including disabled ones, delete some first", name, total))
	}

	return nil
}
