package wire

import (
	"encoding/json"
	"errors"
	"time"
	"unicode"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// Bundle is a bundle with the content of the version apps install.
type Bundle struct {
	ID          string `json:"id"`
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Listing is how it's shown: the listing Kite approved while it's listed,
	// otherwise the page of a user's bundle. Name and Description are the same
	// as its own.
	Listing *BundleListing `json:"listing"`
	// Listed is whether it's listed in the marketplace.
	Listed bool `json:"listed"`
	// Content is the version in the format of the templates of kite-web.
	Content json.RawMessage `json:"content"`
	// Official bundles are made by Kite, the others by users.
	Official bool `json:"official"`
	// Yours is whether the current user made it.
	Yours bool `json:"yours"`
	// Owner is the user who made it, null for official bundles.
	Owner *BundleOwner `json:"owner"`
	// VersionNotes are what changed in the last versions, newest first.
	VersionNotes []BundleVersionNotes `json:"version_notes"`
	// InstallCount is how many apps added it.
	InstallCount int `json:"install_count"`
	// StarCount is how many users starred it.
	StarCount int `json:"star_count"`
	// CreatedAt is when it was made, which the marketplace shows the newest
	// by.
	CreatedAt time.Time `json:"created_at"`
}

type BundleVersionNotes struct {
	Version   int       `json:"version"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

// BundleOwner is the user who made a bundle, as anyone can see it on Discord.
type BundleOwner struct {
	DisplayName     string      `json:"display_name"`
	DiscordID       string      `json:"discord_id"`
	DiscordUsername string      `json:"discord_username"`
	DiscordAvatar   null.String `json:"discord_avatar"`
}

func BundleOwnerToWire(user *model.User) *BundleOwner {
	return &BundleOwner{
		DisplayName:     user.DisplayName,
		DiscordID:       user.DiscordID,
		DiscordUsername: user.DiscordUsername,
		DiscordAvatar:   user.DiscordAvatar,
	}
}

// BundleListing is how a bundle is shown on its page and in the marketplace.
// Tags are lowercase words on what it does, like roblox or economy.
type BundleListing struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Icon        string      `json:"icon"`
	Category    string      `json:"category"`
	Tags        []string    `json:"tags"`
	About       []string    `json:"about"`
	FAQ         []BundleFAQ `json:"faq"`
	Setup       []string    `json:"setup"`
}

type BundleFAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type BundleListResponse = []*Bundle

type BundleGetResponse = Bundle

func BundleToWire(bundle *model.Bundle) *Bundle {
	res := &Bundle{
		ID:           bundle.ID,
		Version:      bundle.Version,
		Name:         bundle.Name,
		Description:  bundle.Description,
		Content:      bundle.Content,
		Listed:       bundle.Listed,
		Official:     !bundle.OwnerUserID.Valid,
		VersionNotes: make([]BundleVersionNotes, len(bundle.VersionNotes)),
		InstallCount: bundle.InstallCount,
		StarCount:    bundle.StarCount,
		CreatedAt:    bundle.CreatedAt,
	}
	for i, n := range bundle.VersionNotes {
		res.VersionNotes[i] = BundleVersionNotes{Version: n.Version, Notes: n.Notes, CreatedAt: n.CreatedAt}
	}

	if listing := bundle.Listing; listing != nil {
		l := BundleListingToWire(*listing)
		res.Listing = &l
	}

	return res
}

// UserBundle is a bundle the user made, without its content.
type UserBundle struct {
	ID          string `json:"id"`
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// SourceAppID is the app new versions are published from, null if it was
	// deleted.
	SourceAppID null.String `json:"source_app_id"`
	// ItemCount is how many commands, event listeners, messages and variables
	// the version has.
	ItemCount int `json:"item_count"`
	// InstallCount is how many apps added it, also ones not from the
	// marketplace.
	InstallCount int `json:"install_count"`
	// ListedVersion is the version apps install from the marketplace, null if
	// it isn't listed.
	ListedVersion null.Int `json:"listed_version"`
	// ApprovedVersion is the version a review approved last, which apps
	// install also after it leaves the marketplace, null if none was.
	ApprovedVersion null.Int `json:"approved_version"`
	// Page is its listing as the user last wrote it. While it's listed, the
	// marketplace shows the one Kite approved last.
	Page BundleListing `json:"page"`
	// LatestReview is the newest review for the marketplace, null if it was
	// never submitted.
	LatestReview *BundleReview `json:"latest_review"`
	// RemovedReason is why Kite took the bundle down, null if it didn't.
	RemovedReason null.String `json:"removed_reason"`
}

type UserBundleListResponse = []*UserBundle

// BundleContentRequest is what a version of a bundle is made of: items of the
// app, with the message templates and stored variables they use.
type BundleContentRequest struct {
	// SharedParameterKeys are the text parameters whose values the owner
	// shares. Apps that add the bundle start with them.
	SharedParameterKeys []string `json:"shared_parameter_keys,omitempty"`
	BundleExportRequest `tstype:",extends"`
}

func (req BundleContentRequest) Validate() error {
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.SharedParameterKeys, validation.Length(0, 500), validation.Each(validation.Length(1, 100))),
	); err != nil {
		return err
	}
	return req.BundleExportRequest.Validate()
}

// validateBundleName rejects emoji and other symbols in names, which the
// marketplace shows next to the icon of the bundle. kite-web checks the same.
func validateBundleName(value any) error {
	for _, r := range value.(string) {
		if unicode.Is(unicode.So, r) {
			return errors.New("can't have emojis or symbols, pick an icon for the marketplace instead")
		}
	}
	return nil
}

// BundleCreateRequest makes a bundle of the user from items of the app.
type BundleCreateRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description"`
	BundleContentRequest `tstype:",extends"`
}

func (req BundleCreateRequest) Validate() error {
	page := BundleListing{Name: req.Name, Description: req.Description}
	if err := page.validate(false); err != nil {
		return err
	}
	return req.BundleContentRequest.Validate()
}

type BundleCreateResponse = UserBundle

// BundleVersionCreateRequest publishes items of the app as the new version of
// a bundle. They can be different ones than before. Its name and description
// are part of its page, which changes without a new version.
type BundleVersionCreateRequest struct {
	BundleContentRequest `tstype:",extends"`
	// Notes are what changed, for the apps that update to it.
	Notes string `json:"notes"`
}

func (req BundleVersionCreateRequest) Validate() error {
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Notes, validation.Length(0, 1000)),
	); err != nil {
		return err
	}
	return req.BundleContentRequest.Validate()
}

type BundleVersionCreateResponse = UserBundle

// BundleUpdateRequest changes the page of a bundle that isn't in the
// marketplace. Kite reviews changes to one that is, so they're submitted with
// BundleListingSubmitRequest instead.
type BundleUpdateRequest struct {
	BundleListing `tstype:",extends"`
}

func (req BundleUpdateRequest) Validate() error {
	return req.BundleListing.validate(false)
}

type BundleUpdateResponse struct{}

type BundleDeleteResponse struct{}

type BundleListingDeleteResponse struct{}

func UserBundleToWire(bundle *model.Bundle) *UserBundle {
	res := &UserBundle{
		ID:           bundle.ID,
		Version:      bundle.Version,
		Name:         bundle.Name,
		Description:  bundle.Description,
		SourceAppID:  bundle.SourceAppID,
		ItemCount:    bundle.ItemCount,
		InstallCount: bundle.InstallCount,
		Page:         BundleListingToWire(bundle.Page),
		LatestReview: BundleReviewToWire(bundle.LatestReview),
	}
	if bundle.RemovedAt.Valid {
		res.RemovedReason = null.StringFrom(bundle.RemovedReason)
	}
	if v := bundle.ListedVersion(); v > 0 {
		res.ListedVersion = null.IntFrom(int64(v))
	}
	if bundle.ApprovedVersion > 0 {
		res.ApprovedVersion = null.IntFrom(int64(bundle.ApprovedVersion))
	}
	return res
}
