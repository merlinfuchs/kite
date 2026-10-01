package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

type MarketplaceListingStatus string

const (
	// MarketplaceListingStatusPending listings wait for a moderator and are
	// only visible to their author and moderators.
	MarketplaceListingStatusPending  MarketplaceListingStatus = "pending"
	MarketplaceListingStatusApproved MarketplaceListingStatus = "approved"
	MarketplaceListingStatusRejected MarketplaceListingStatus = "rejected"
	// MarketplaceListingStatusRemoved listings were approved before and taken
	// down by a moderator.
	MarketplaceListingStatusRemoved MarketplaceListingStatus = "removed"
)

type MarketplaceListingItemType string

const (
	MarketplaceListingItemTypeCommand       MarketplaceListingItemType = "command"
	MarketplaceListingItemTypeEventListener MarketplaceListingItemType = "event_listener"
	MarketplaceListingItemTypeMessage       MarketplaceListingItemType = "message"
)

type MarketplaceListing struct {
	ID                 string
	Name               string
	Description        string
	AuthorUserID       string
	SourceAppID        null.String
	Status             MarketplaceListingStatus
	Items              []MarketplaceListingItem
	CommandCount       int
	EventListenerCount int
	MessageCount       int
	BlockTypes         []string
	ImportCount        int
	ReviewNote         null.String
	ReviewedByUserID   null.String
	ReviewedAt         null.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// IsModule is true for listings with more than one item.
func (l *MarketplaceListing) IsModule() bool {
	return l.CommandCount+l.EventListenerCount+l.MessageCount > 1
}

// MarketplaceListingItem is one command, event listener or message template of
// a listing. For commands and event listeners Name and Description are taken
// from the compiled flow, not from the author.
type MarketplaceListingItem struct {
	Type        MarketplaceListingItemType `json:"type"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	// Source is the event source, only set for event listeners.
	Source string `json:"source,omitempty"`
	// FlowSource is the flow of commands and event listeners.
	FlowSource flow.FlowData `json:"flow_source"`

	// SourceID is the message template's ID in the author's app. Blocks in
	// the same listing reference the template by it, so they can be pointed
	// at the imported copy.
	SourceID string `json:"source_id,omitempty"`
	// MessageData and MessageFlowSources are only set for message templates.
	MessageData        *message.MessageData     `json:"message_data,omitempty"`
	MessageFlowSources map[string]flow.FlowData `json:"message_flow_sources,omitempty"`
}

type MarketplaceListingWithAuthor struct {
	MarketplaceListing
	Author *User
}

type MarketplaceReport struct {
	ID               string
	ListingID        string
	ReporterUserID   string
	Reason           string
	CreatedAt        time.Time
	ResolvedAt       null.Time
	ResolvedByUserID null.String
}

type MarketplaceReportWithReporter struct {
	MarketplaceReport
	Reporter *User
}

type MarketplaceModerator struct {
	DiscordUserID string
	AddedByUserID null.String
	CreatedAt     time.Time
}
