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
	// Version is the version apps install.
	Version int
	// Content is that version, in the format of internal/core/bundle.
	Content json.RawMessage
	// Listing is nil for bundles that weren't published.
	Listing   *BundleListing
	CreatedAt time.Time
	UpdatedAt time.Time
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
