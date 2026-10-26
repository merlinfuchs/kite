package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// MaxFlowVersions is how many versions are kept per flow.
const MaxFlowVersions = 50

type FlowVersionTargetType string

const (
	FlowVersionTargetTypeCommand       FlowVersionTargetType = "command"
	FlowVersionTargetTypeEventListener FlowVersionTargetType = "event_listener"
)

// FlowVersionSource is what made a flow version.
type FlowVersionSource string

const (
	// The flow from before its first version, made by the first change.
	FlowVersionSourceInitial  FlowVersionSource = "initial"
	FlowVersionSourceEditor   FlowVersionSource = "editor"
	FlowVersionSourceAutoSave FlowVersionSource = "auto_save"
	// A template or bundle update, or a change of its parameters.
	FlowVersionSourceBundle FlowVersionSource = "bundle"
	// The last flow of a deleted command or event listener.
	FlowVersionSourceDelete FlowVersionSource = "delete"
)

// FlowVersion is a saved state of the flow of a command or event listener.
type FlowVersion struct {
	ID         string
	AppID      string
	TargetType FlowVersionTargetType
	TargetID   string
	Source     FlowVersionSource
	// Null for initial versions.
	CreatorUserID null.String
	// Only set in lists.
	CreatorDisplayName null.String
	// Empty in lists, which only load it for a single version.
	FlowSource flow.FlowData
	CreatedAt  time.Time
}
