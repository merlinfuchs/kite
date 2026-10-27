package model

import (
	"encoding/json"
	"time"

	"gopkg.in/guregu/null.v4"
)

// Bundle is something apps can install, for now an official template.
type Bundle struct {
	ID string
	// Name and Description are how the bundle is shown, like Listing.
	Name        string
	Description string
	// OwnerUserID is null for official bundles.
	OwnerUserID null.String
	// SourceAppID is the app a user's bundle is exported from, null for
	// official bundles and ones whose app was deleted.
	SourceAppID null.String
	// RemovedAt is when Kite took a user's bundle down for breaking the
	// rules, with the reason its owner sees. It's final.
	RemovedAt     null.Time
	RemovedReason string
	// Version is the version of Content. Loaders with
	// the content load InstallVersion, the others the newest. Use
	// InstallVersion and LatestVersion to know which version is which.
	Version int
	// LatestVersion is the newest version, which new versions are published
	// after. It's newer than the install version while a new version is in
	// review.
	LatestVersion int
	// ApprovedVersion is the version of a user's bundle a review approved
	// last, or 0 if none was.
	ApprovedVersion int
	// Listed is whether the bundle is listed in the marketplace, which
	// official ones always are, and users' ones at their approved version.
	Listed bool
	// InReview is whether a review of a user's bundle is pending, set when
	// it's loaded on its own.
	InReview bool
	// Content is the content of Version, in the format of
	// internal/core/bundle.
	Content json.RawMessage
	// ItemCount is how many items the version has, set when the bundle is
	// loaded without its content.
	ItemCount int
	// InstallCount is how many apps added the bundle, set when the bundles of
	// its owner are listed.
	InstallCount int
	// Notes are what changed in the version, set when it's published.
	Notes string
	// VersionNotes are what changed in the kept versions, newest first.
	VersionNotes []BundleVersionNotes
	// Listing is how the bundle is shown: the listing a review approved while
	// it's listed, otherwise the page of a user's bundle. It's nil for
	// official bundles that aren't listed.
	Listing *BundleListing
	// Page is how its owner last wrote the listing of a user's bundle, which
	// its link shows, and the marketplace once a review approves it.
	Page BundleListing
	// LatestReview is the newest review of a user's bundle, set when the
	// owner's bundles are loaded. It's nil if it was never submitted.
	LatestReview *BundleReview
	// Owner is who made a user's bundle, set when it's loaded for others to
	// see who made it, like the admins.
	Owner     *User
	CreatedAt time.Time
	UpdatedAt time.Time
}

// InstallVersion is the version apps install: the approved one of a user's
// bundle once a review approved one, also after it left the marketplace, so
// apps only get versions Kite reviewed. Otherwise it's the newest.
func (b *Bundle) InstallVersion() int {
	if b.ApprovedVersion > 0 {
		return b.ApprovedVersion
	}
	return b.LatestVersion
}

// ListedVersion is the version of a user's bundle listed in the marketplace,
// or 0 if it isn't. Official bundles have no approved version.
func (b *Bundle) ListedVersion() int {
	if !b.Listed {
		return 0
	}
	return b.ApprovedVersion
}

// ShowListing shows the bundle as the listing, with its name and description.
func (b *Bundle) ShowListing(l *BundleListing) {
	b.Listing = l
	b.Name = l.Name
	b.Description = l.Description
}

// OwnedBy is whether the user made the bundle.
func (b *Bundle) OwnedBy(userID string) bool {
	return b.OwnerUserID.Valid && b.OwnerUserID.String == userID
}

// VisibleTo is whether the user can see and install the bundle. Anyone with
// its link can, unless Kite took it down, then only its owner.
func (b *Bundle) VisibleTo(userID string) bool {
	return !b.OwnerUserID.Valid || b.OwnedBy(userID) || !b.RemovedAt.Valid
}

// BundleVersionNotes are what changed in a version of a bundle.
type BundleVersionNotes struct {
	Version   int
	Notes     string
	CreatedAt time.Time
}

// BundleListing is how a bundle is shown on its page and in the marketplace.
// Official bundles always have one, users' bundles write theirs and need a
// review to list it. It can change without a new version.
type BundleListing struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Icon        string      `json:"icon"`
	Category    string      `json:"category"`
	About       []string    `json:"about"`
	FAQ         []BundleFAQ `json:"faq"`
	Setup       []string    `json:"setup"`
	Position    int         `json:"-"`
	UpdatedAt   time.Time   `json:"-"`
}

// Complete is whether the listing has what the marketplace needs: an icon, a
// category and how it works.
func (l BundleListing) Complete() bool {
	return l.Icon != "" && l.Category != "" && len(l.About) > 0
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
	// BundleReviewStatusWithdrawn is a review that was pending or rejected
	// when the bundle left the marketplace. A rejection keeps its reason.
	BundleReviewStatusWithdrawn BundleReviewStatus = "withdrawn"
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

// PendingBundleReview is a review waiting for a decision, with the bundle as
// the reviewed version and the bundle's owner.
type PendingBundleReview struct {
	Review *BundleReview
	Bundle *Bundle
}

type BundleReportCategory string

const (
	BundleReportCategoryHarmful       BundleReportCategory = "harmful"
	BundleReportCategoryScam          BundleReportCategory = "scam"
	BundleReportCategoryInappropriate BundleReportCategory = "inappropriate"
	BundleReportCategorySpam          BundleReportCategory = "spam"
	BundleReportCategoryOther         BundleReportCategory = "other"
)

// BundleReport is a report that a user's bundle breaks the rules.
type BundleReport struct {
	ID             string
	BundleID       string
	ReporterUserID string
	Category       BundleReportCategory
	Details        string
	CreatedAt      time.Time
	ResolvedAt     null.Time
	ResolverUserID null.String
}

// OpenBundleReport is a report waiting for a decision, with its bundle.
type OpenBundleReport struct {
	Report *BundleReport
	Bundle *Bundle
}
