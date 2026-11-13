package wire

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// EditorSessionUpdateRequest reports how long an editor of a command or event
// listener has been in use so far.
type EditorSessionUpdateRequest struct {
	ActiveSeconds int `json:"active_seconds"`
}

func (req EditorSessionUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ActiveSeconds, validation.Min(0)),
	)
}

type EditorSessionUpdateResponse = Empty

// EditorUsage is how much commands and event listeners were worked on in the
// editor.
type EditorUsage struct {
	ActiveSeconds int       `json:"active_seconds"`
	SaveCount     int       `json:"save_count"`
	LastSavedAt   null.Time `json:"last_saved_at"`
}

// EditorUsageGetResponse has the usage of the last 30 days, for the whole app
// and by the ID of each command and event listener that was worked on.
type EditorUsageGetResponse struct {
	Total          EditorUsage             `json:"total"`
	Commands       map[string]*EditorUsage `json:"commands"`
	EventListeners map[string]*EditorUsage `json:"event_listeners"`
}

func EditorUsageToWire(usage *model.EditorUsage) *EditorUsage {
	return &EditorUsage{
		ActiveSeconds: usage.ActiveSeconds,
		SaveCount:     usage.SaveCount,
		LastSavedAt:   usage.LastSavedAt,
	}
}

// Add adds other to the usage.
func (u *EditorUsage) Add(other *EditorUsage) {
	u.ActiveSeconds += other.ActiveSeconds
	u.SaveCount += other.SaveCount
	if other.LastSavedAt.Valid && (!u.LastSavedAt.Valid || other.LastSavedAt.Time.After(u.LastSavedAt.Time)) {
		u.LastSavedAt = other.LastSavedAt
	}
}
