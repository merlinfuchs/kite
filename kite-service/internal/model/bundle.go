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
	// How the bundle is shown, which can change without a new version.
	Icon      string
	Category  string
	About     []string
	FAQ       []BundleFAQ
	Setup     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BundleFAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// BundleVersion is the content of a version of a bundle, which never changes
// for bundles of users.
type BundleVersion struct {
	BundleID  string
	Version   int
	Content   json.RawMessage
	CreatedAt time.Time
}
