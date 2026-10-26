package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// BundleReportCreateRequest reports that a bundle breaks the rules.
type BundleReportCreateRequest struct {
	// Category is harmful, scam, inappropriate, spam or other.
	Category string `json:"category"`
	Details  string `json:"details"`
}

func (req BundleReportCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Category, validation.Required, validation.In(
			string(model.BundleReportCategoryHarmful),
			string(model.BundleReportCategoryScam),
			string(model.BundleReportCategoryInappropriate),
			string(model.BundleReportCategorySpam),
			string(model.BundleReportCategoryOther),
		)),
		validation.Field(&req.Details, validation.Required, validation.Length(1, 1000)),
	)
}

type BundleReportCreateResponse struct{}

// OpenBundleReport is a report waiting for a decision, for the admins.
type OpenBundleReport struct {
	ID                string       `json:"id"`
	BundleID          string       `json:"bundle_id"`
	BundleName        string       `json:"bundle_name"`
	BundleDescription string       `json:"bundle_description"`
	Owner             *BundleOwner `json:"owner"`
	Reporter          *BundleOwner `json:"reporter"`
	Category          string       `json:"category"`
	Details           string       `json:"details"`
	CreatedAt         time.Time    `json:"created_at"`
}

type OpenBundleReportListResponse = []*OpenBundleReport

type BundleReportDismissResponse struct{}

// BundleRemoveRequest takes a bundle down for breaking the rules.
type BundleRemoveRequest struct {
	// Reason is shown to the bundle's owner.
	Reason string `json:"reason"`
}

func (req BundleRemoveRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Reason, validation.Required, validation.Length(1, 1000)),
	)
}

type BundleRemoveResponse struct{}
