package model

type Plan struct {
	ID          string  `toml:"id" validate:"required"`
	Title       string  `toml:"title" validate:"required"`
	Description string  `toml:"description" validate:"required"`
	Price       float32 `toml:"price" validate:"required"`
	Default     bool    `toml:"default"`
	Popular     bool    `toml:"popular"`
	Hidden      bool    `toml:"hidden"`

	LemonSqueezyProductID string `toml:"lemonsqueezy_product_id"`
	LemonSqueezyVariantID string `toml:"lemonsqueezy_variant_id"`

	FeatureMaxCollaborators     int  `toml:"feature_max_collaborators"`
	FeatureUsageCreditsPerMonth int  `toml:"feature_usage_credits_per_month"`
	FeatureMaxGuilds            int  `toml:"feature_max_guilds"`
	FeaturePrioritySupport      bool `toml:"feature_priority_support"`
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
