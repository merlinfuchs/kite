package wire

import (
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// FlowVersion is a saved state of a flow. Lists leave out the flow itself.
type FlowVersion struct {
	ID string `json:"id"`
	// What made the version: initial, editor, auto_save, bundle or delete.
	Source             string      `json:"source"`
	CreatorUserID      null.String `json:"creator_user_id"`
	CreatorDisplayName null.String `json:"creator_display_name"`
	CreatedAt          time.Time   `json:"created_at"`
}

type FlowVersionListResponse = []*FlowVersion

type FlowVersionGetResponse struct {
	ID            string        `json:"id"`
	Source        string        `json:"source"`
	CreatorUserID null.String   `json:"creator_user_id"`
	FlowSource    flow.FlowData `json:"flow_source"`
	CreatedAt     time.Time     `json:"created_at"`
}

func FlowVersionToWire(version *model.FlowVersion) *FlowVersion {
	if version == nil {
		return nil
	}

	return &FlowVersion{
		ID:                 version.ID,
		Source:             string(version.Source),
		CreatorUserID:      version.CreatorUserID,
		CreatorDisplayName: version.CreatorDisplayName,
		CreatedAt:          version.CreatedAt,
	}
}
