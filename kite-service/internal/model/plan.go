package model

import "time"

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
	FeatureRotatingStatus       bool

	FeatureMaxScheduledEventListeners int
	FeatureMinScheduleIntervalSeconds int
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
		RotatingStatus:       p.FeatureRotatingStatus,

		MaxScheduledEventListeners: p.FeatureMaxScheduledEventListeners,
		MinScheduleIntervalSeconds: p.FeatureMinScheduleIntervalSeconds,
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
	RotatingStatus       bool

	MaxScheduledEventListeners int
	MinScheduleIntervalSeconds int
}

// DefaultMinScheduleInterval applies to plans that don't set a minimum, so a
// plan config missing the field can't accidentally allow per-second schedules.
const DefaultMinScheduleInterval = 5 * time.Minute

func (f Features) MinScheduleInterval() time.Duration {
	if f.MinScheduleIntervalSeconds <= 0 {
		return DefaultMinScheduleInterval
	}
	return time.Duration(f.MinScheduleIntervalSeconds) * time.Second
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
		RotatingStatus:       f.RotatingStatus || other.RotatingStatus,

		MaxScheduledEventListeners: max(f.MaxScheduledEventListeners, other.MaxScheduledEventListeners),
		// A shorter interval is the better one, unlike every other field.
		MinScheduleIntervalSeconds: minSet(f.MinScheduleIntervalSeconds, other.MinScheduleIntervalSeconds),
	}
}

// minSet returns the smaller of a and b, ignoring either if it's unset.
func minSet(a, b int) int {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	return min(a, b)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
