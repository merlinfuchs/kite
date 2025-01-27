package wire

import (
	"time"
)

type Entitlement struct {
	ID           string                `json:"id"`
	Default      bool                  `json:"default"`
	Subscription *BillingSubscription  `json:"subscription"`
	FeatureSet   EntitlementFeatureSet `json:"feature_set"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

type EntitlementFeatureSet struct {
	UsageCreditsPerMonth int `json:"usage_credits_per_month"`
	MaxCollaborators     int `json:"max_collaborators"`
}

type EntitlementListResponse = []*Entitlement

type EntitlementFeaturesGetResponse = EntitlementFeatureSet
