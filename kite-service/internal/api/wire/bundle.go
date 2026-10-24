package wire

import (
	"encoding/json"

	"github.com/kitecloud/kite/kite-service/internal/model"
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
