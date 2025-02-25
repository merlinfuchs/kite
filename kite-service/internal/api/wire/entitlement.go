package wire

type EntitlementFeatures struct {
	UsageCreditsPerMonth int `json:"usage_credits_per_month"`
	MaxCollaborators     int `json:"max_collaborators"`
}

type EntitlementFeaturesGetResponse = EntitlementFeatures
