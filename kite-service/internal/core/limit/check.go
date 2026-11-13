package limit

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// TotalFactor is how many times its limit an app can have of commands or
// event listeners, counting the disabled ones. Only enabled ones count
// against the limit, so apps can keep some disabled instead of deleting them.
const TotalFactor = 3

// EventListenerLimit returns the limit of the event listeners of the source,
// 0 when there is none, and their name in the dashboard. Each source has its
// own limit.
func EventListenerLimit(f model.Features, source model.EventSource) (int, string) {
	switch source {
	case model.EventSourceSchedule:
		return f.MaxScheduledEventListeners, "scheduled tasks"
	case model.EventSourceWebhook:
		return f.MaxWebhookEventListeners, "webhooks"
	default:
		return f.MaxEventListeners, "event listeners"
	}
}

// Check checks that an app with counts can have added more items. name is
// the plural the error uses, its message is for the user.
func Check(name string, limit int, counts store.ItemCounts, added store.ItemCounts) error {
	// Apps over their limit after a downgrade can still add disabled items.
	if added.Enabled > 0 && counts.Enabled+added.Enabled > limit {
		return fmt.Errorf("maximum number of enabled %s (%d) reached, disable or delete some first", name, limit)
	}

	if total := limit * TotalFactor; counts.Total+added.Total > total {
		return fmt.Errorf("maximum number of %s (%d) reached, including disabled ones, delete some first", name, total)
	}

	return nil
}
