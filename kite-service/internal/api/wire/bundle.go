package wire

import (
	"encoding/json"

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
	// Listing is null for bundles that weren't published.
	Listing *BundleListing `json:"listing"`
	// Content is the version in the format of the templates of kite-web.
	Content json.RawMessage `json:"content"`
	// Official bundles are made by Kite, the others by users.
	Official bool `json:"official"`
	// Yours is whether the current user made it.
	Yours bool `json:"yours"`
}

// BundleListing is how a published bundle is shown.
type BundleListing struct {
	Icon     string      `json:"icon"`
	Category string      `json:"category"`
	About    []string    `json:"about"`
	FAQ      []BundleFAQ `json:"faq"`
	Setup    []string    `json:"setup"`
}

type BundleFAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type BundleListResponse = []*Bundle

type BundleGetResponse = Bundle

func BundleToWire(bundle *model.Bundle) *Bundle {
	res := &Bundle{
		ID:          bundle.ID,
		Version:     bundle.Version,
		Name:        bundle.Name,
		Description: bundle.Description,
		Content:     bundle.Content,
		Official:    !bundle.OwnerUserID.Valid,
	}

	if listing := bundle.Listing; listing != nil {
		faq := make([]BundleFAQ, len(listing.FAQ))
		for i, f := range listing.FAQ {
			faq[i] = BundleFAQ{Question: f.Question, Answer: f.Answer}
		}
		res.Listing = &BundleListing{
			Icon:     listing.Icon,
			Category: listing.Category,
			About:    listing.About,
			FAQ:      faq,
			Setup:    listing.Setup,
		}
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
}

type UserBundleListResponse = []*UserBundle

// BundleCreateRequest makes a bundle of the user from items of the app. The
// message templates and stored variables they use are part of it.
type BundleCreateRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	BundleExportRequest `tstype:",extends"`
}

func (req BundleCreateRequest) Validate() error {
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Description, validation.Length(0, 1000)),
	); err != nil {
		return err
	}
	return req.BundleExportRequest.Validate()
}

type BundleCreateResponse = UserBundle

// BundleVersionCreateRequest publishes items of the app as the new version of
// a bundle. They can be different ones than before.
type BundleVersionCreateRequest = BundleCreateRequest

type BundleVersionCreateResponse = UserBundle

type BundleDeleteResponse struct{}

func UserBundleToWire(bundle *model.Bundle) *UserBundle {
	return &UserBundle{
		ID:          bundle.ID,
		Version:     bundle.Version,
		Name:        bundle.Name,
		Description: bundle.Description,
		SourceAppID: bundle.SourceAppID,
		ItemCount:   bundle.ItemCount,
	}
}
