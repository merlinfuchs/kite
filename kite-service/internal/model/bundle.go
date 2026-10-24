package model

import (
	"encoding/json"
	"time"

	"gopkg.in/guregu/null.v4"
)

// Bundle is something apps can install, for now an official template.
type Bundle struct {
	ID          string
	Name        string
	Description string
	// OwnerUserID is null for official bundles.
	OwnerUserID null.String
	// SourceAppID is the app a user's bundle is exported from, null for
	// official bundles and ones whose app was deleted.
	SourceAppID null.String
	// Version is the version apps install.
	Version int
	// Content is that version, in the format of internal/core/bundle.
	Content json.RawMessage
	// ItemCount is how many items the version has, set when the bundle is
	// loaded without its content.
	ItemCount int
	// Listing is nil for bundles that weren't published.
	Listing   *BundleListing
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OwnedBy is whether the user made the bundle.
func (b *Bundle) OwnedBy(userID string) bool {
	return b.OwnerUserID.Valid && b.OwnerUserID.String == userID
}

// VisibleTo is whether the user can see and install the bundle. Users'
// bundles are only visible to them.
func (b *Bundle) VisibleTo(userID string) bool {
	return !b.OwnerUserID.Valid || b.OwnedBy(userID)
}

// BundleListing is how a published bundle is shown, on its page and later in
// the marketplace. It can change without a new version.
type BundleListing struct {
	Icon      string      `json:"icon"`
	Category  string      `json:"category"`
	About     []string    `json:"about"`
	FAQ       []BundleFAQ `json:"faq"`
	Setup     []string    `json:"setup"`
	Position  int         `json:"-"`
	UpdatedAt time.Time   `json:"-"`
}

type BundleFAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
