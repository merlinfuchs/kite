package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
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
	BlockTypes         []string
	ImportCount        int
	ReviewNote         null.String
	ReviewedByUserID   null.String
	ReviewedAt         null.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// IsModule is true for listings with more than one command or event listener.
func (l *MarketplaceListing) IsModule() bool {
	return l.CommandCount+l.EventListenerCount > 1
}

// MarketplaceListingItem is one command or event listener of a listing. Name
// and Description are taken from the compiled flow, not from the author.
type MarketplaceListingItem struct {
	Type        MarketplaceListingItemType `json:"type"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	// Source is the event source, only set for event listeners.
	Source     string        `json:"source,omitempty"`
	FlowSource flow.FlowData `json:"flow_source"`
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
