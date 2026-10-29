package wire

import (
	"encoding/json"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// ItemVersion is a saved state of a command, event listener or message,
// without its content.
type ItemVersion struct {
	ID string `json:"id"`
	// What made the version: initial, editor, api_token, bundle or delete.
	Source             string      `json:"source"`
	CreatorUserID      null.String `json:"creator_user_id"`
	CreatorDisplayName null.String `json:"creator_display_name"`
	CreatedAt          time.Time   `json:"created_at"`
	// When later changes merged into the version last updated it, if any.
	UpdatedAt null.Time `json:"updated_at"`
}

type ItemVersionListResponse = []*ItemVersion

type ItemVersionGetResponse struct {
	ID            string      `json:"id"`
	Source        string      `json:"source"`
	CreatorUserID null.String `json:"creator_user_id"`
	// The flow of a command or event listener, or the data and flow_sources
	// of a message.
	Content   json.RawMessage `json:"content"`
	CreatedAt time.Time       `json:"created_at"`
}

func ItemVersionToWire(version *model.ItemVersion) *ItemVersion {
	if version == nil {
		return nil
	}

	return &ItemVersion{
		ID:                 version.ID,
		Source:             string(version.Source),
		CreatorUserID:      version.CreatorUserID,
		CreatorDisplayName: version.CreatorDisplayName,
		CreatedAt:          version.CreatedAt,
		UpdatedAt:          version.UpdatedAt,
	}
}
