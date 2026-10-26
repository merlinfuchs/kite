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
	// PublishedAt is when the user published it, so anyone with its link can
	// see and install it.
	PublishedAt null.Time
	// Version is the version apps install, the listed one of a listed user's
	// bundle.
	Version int
	// LatestVersion is the newest version, which new versions are published
	// after. It's newer than Version while a new version is in review.
	LatestVersion int
	// Content is that version, in the format of internal/core/bundle.
	Content json.RawMessage
	// ItemCount is how many items the version has, set when the bundle is
	// loaded without its content.
	ItemCount int
	// Notes are what changed in the version, set when it's published.
	Notes string
	// VersionNotes are what changed in the kept versions, newest first.
	VersionNotes []BundleVersionNotes
	// Listing is nil for bundles that aren't listed.
	Listing *BundleListing
	// LatestReview is the newest review of a user's bundle, set when the
	// owner's bundles are loaded. It's nil if it was never submitted.
	LatestReview *BundleReview
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// OwnedBy is whether the user made the bundle.
func (b *Bundle) OwnedBy(userID string) bool {
	return b.OwnerUserID.Valid && b.OwnerUserID.String == userID
}

// VisibleTo is whether the user can see and install the bundle. Users'
// bundles are only visible to others once they're published.
func (b *Bundle) VisibleTo(userID string) bool {
	return !b.OwnerUserID.Valid || b.OwnedBy(userID) || b.PublishedAt.Valid
}

// BundleVersionNotes are what changed in a version of a bundle.
type BundleVersionNotes struct {
	Version   int
	Notes     string
	CreatedAt time.Time
}

// BundleListing is how a listed bundle is shown in the marketplace. Official
// bundles always have one, users' bundles once a review approved them.
type BundleListing struct {
	Icon     string      `json:"icon"`
	Category string      `json:"category"`
	About    []string    `json:"about"`
	FAQ      []BundleFAQ `json:"faq"`
	Setup    []string    `json:"setup"`
	Position int         `json:"-"`
	// ListedVersion is the version of a user's bundle apps install, the last
	// one a review approved. It's 0 for official bundles, which install their
	// newest version.
	ListedVersion int       `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

type BundleFAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type BundleReviewStatus string

const (
	BundleReviewStatusPending  BundleReviewStatus = "pending"
	BundleReviewStatusApproved BundleReviewStatus = "approved"
	BundleReviewStatusRejected BundleReviewStatus = "rejected"
)

// BundleReview is a review of a version of a user's bundle before it's listed
// in the marketplace, either because it was submitted or because a listed
// bundle got a new version.
type BundleReview struct {
	ID       string
	BundleID string
	Version  int
	// Listing is how the bundle is shown once it's approved.
	Listing BundleListing
	Status  BundleReviewStatus
	// Reason is why it was rejected, which the owner sees.
	Reason         string
	ReviewerUserID null.String
	CreatedAt      time.Time
	DecidedAt      null.Time
}

// PendingBundleReview is a review waiting for a decision, with the bundle and
// the content of the reviewed version.
type PendingBundleReview struct {
	Review *BundleReview
	Bundle *Bundle
}
