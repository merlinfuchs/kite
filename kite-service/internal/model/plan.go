package model

type Plan struct {
	ID          string
	Title       string
	Description string
	Price       float32
	Default     bool
	Popular     bool
	Hidden      bool

	LemonSqueezyProductID string
	LemonSqueezyVariantID string

	DiscordRoleID string

	FeatureMaxCollaborators     int
	FeatureUsageCreditsPerMonth int
	FeatureMaxGuilds            int
	FeatureMaxCommands          int
	FeatureMaxVariables         int
	FeatureMaxMessages          int
	FeatureMaxEventListeners    int
	FeaturePrioritySupport      bool
}

func (p Plan) Features() Features {
	return Features{
		MaxCollaborators:     p.FeatureMaxCollaborators,
		UsageCreditsPerMonth: p.FeatureUsageCreditsPerMonth,
		MaxGuilds:            p.FeatureMaxGuilds,
		MaxCommands:          p.FeatureMaxCommands,
		MaxVariables:         p.FeatureMaxVariables,
		MaxMessages:          p.FeatureMaxMessages,
		MaxEventListeners:    p.FeatureMaxEventListeners,
		PrioritySupport:      p.FeaturePrioritySupport,
	}
}

type Features struct {
	MaxCollaborators     int
	UsageCreditsPerMonth int
	MaxGuilds            int
	MaxCommands          int
	MaxVariables         int
	MaxMessages          int
	MaxEventListeners    int
	PrioritySupport      bool
}

func (f Features) Merge(other Features) Features {
	return Features{
		MaxCollaborators:     max(f.MaxCollaborators, other.MaxCollaborators),
		UsageCreditsPerMonth: max(f.UsageCreditsPerMonth, other.UsageCreditsPerMonth),
		MaxGuilds:            max(f.MaxGuilds, other.MaxGuilds),
		MaxCommands:          max(f.MaxCommands, other.MaxCommands),
		MaxVariables:         max(f.MaxVariables, other.MaxVariables),
		MaxMessages:          max(f.MaxMessages, other.MaxMessages),
		MaxEventListeners:    max(f.MaxEventListeners, other.MaxEventListeners),
		PrioritySupport:      f.PrioritySupport || other.PrioritySupport,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
