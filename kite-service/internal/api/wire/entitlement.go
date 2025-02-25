package wire

type EntitlementFeatures struct {
	UsageCreditsPerMonth int  `json:"usage_credits_per_month"`
	MaxCollaborators     int  `json:"max_collaborators"`
	MaxGuilds            int  `json:"max_guilds"`
	PrioritySupport      bool `json:"priority_support"`
}

type EntitlementFeaturesGetResponse = EntitlementFeatures
