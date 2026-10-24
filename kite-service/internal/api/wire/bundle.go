package wire

import (
	"encoding/json"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// Bundle is a bundle with the content of the version apps install.
type Bundle struct {
	ID          string      `json:"id"`
	Version     int         `json:"version"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Official    bool        `json:"official"`
	Icon        string      `json:"icon"`
	Category    string      `json:"category"`
	About       []string    `json:"about"`
	FAQ         []BundleFAQ `json:"faq"`
	Setup       []string    `json:"setup"`
	// Content is the version in the format of the templates of kite-web.
	Content json.RawMessage `json:"content"`
}

type BundleFAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type BundleListResponse = []*Bundle

type BundleGetResponse = Bundle

func BundleToWire(bundle *model.Bundle, version *model.BundleVersion) *Bundle {
	faq := make([]BundleFAQ, len(bundle.Listing.FAQ))
	for i, f := range bundle.Listing.FAQ {
		faq[i] = BundleFAQ{Question: f.Question, Answer: f.Answer}
	}

	return &Bundle{
		ID:          bundle.ID,
		Version:     bundle.Version,
		Name:        bundle.Name,
		Description: bundle.Description,
		Official:    !bundle.OwnerUserID.Valid,
		Icon:        bundle.Listing.Icon,
		Category:    bundle.Listing.Category,
		About:       bundle.Listing.About,
		FAQ:         faq,
		Setup:       bundle.Listing.Setup,
		Content:     version.Content,
	}
}
