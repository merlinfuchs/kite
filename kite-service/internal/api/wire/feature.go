package wire

type Features struct {
	MaxCollaborators     int  `json:"max_collaborators"`
	UsageCreditsPerMonth int  `json:"usage_credits_per_month"`
	MaxGuilds            int  `json:"max_guilds"`
	PrioritySupport      bool `json:"priority_support"`
}

type FeaturesGetResponse = Features
