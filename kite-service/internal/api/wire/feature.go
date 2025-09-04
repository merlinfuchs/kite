package wire

type Features struct {
	MaxCollaborators     int  `json:"max_collaborators"`
	UsageCreditsPerMonth int  `json:"usage_credits_per_month"`
	MaxGuilds            int  `json:"max_guilds"`
	MaxCommands          int  `json:"max_commands"`
	MaxVariables         int  `json:"max_variables"`
	MaxMessages          int  `json:"max_messages"`
	MaxEventListeners    int  `json:"max_event_listeners"`
	PrioritySupport      bool `json:"priority_support"`
}

type FeaturesGetResponse = Features
