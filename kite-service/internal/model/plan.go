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
	FeaturePrioritySupport      bool
}

func (p Plan) Features() Features {
	return Features{
		MaxCollaborators:     p.FeatureMaxCollaborators,
		UsageCreditsPerMonth: p.FeatureUsageCreditsPerMonth,
		MaxGuilds:            p.FeatureMaxGuilds,
		PrioritySupport:      p.FeaturePrioritySupport,
	}
}

type Features struct {
	MaxCollaborators     int
	UsageCreditsPerMonth int
	MaxGuilds            int
	PrioritySupport      bool
}

func (f Features) Merge(other Features) Features {
	return Features{
		MaxCollaborators:     max(f.MaxCollaborators, other.MaxCollaborators),
		UsageCreditsPerMonth: max(f.UsageCreditsPerMonth, other.UsageCreditsPerMonth),
		MaxGuilds:            max(f.MaxGuilds, other.MaxGuilds),
		PrioritySupport:      f.PrioritySupport || other.PrioritySupport,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
