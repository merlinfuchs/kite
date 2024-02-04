package dismodel

import "github.com/merlinfuchs/kite/kite-types/null"

type Guild struct {
	ID                          string                          `json:"id"`
	Name                        string                          `json:"name"`
	Icon                        string                          `json:"icon,omitempty"`
	Splash                      string                          `json:"splash,omitempty"`
	DiscoverySplash             string                          `json:"discovery_splash,omitempty"`
	OwnerID                     string                          `json:"owner_id"`
	AFKChannelID                string                          `json:"afk_channel_id,omitempty"`
	AFKTimeout                  int                             `json:"afk_timeout"`
	WidgetEnabled               bool                            `json:"widget_enabled,omitempty"`
	WidgetChannelID             string                          `json:"widget_channel_id,omitempty"`
	VerificationLevel           GuildVerificationLevel          `json:"verification_level"`
	DefaultMessageNotifications GuildDefaultNotificationLevel   `json:"default_message_notifications"`
	ExplicitContentFilter       GuildExplicitContentFilterLevel `json:"explicit_content_filter"`
	Roles                       []Role                          `json:"roles,omitempty"`
	// Emojis                      []Emoji                         `json:"emojis,omitempty"`
	Features                  []string         `json:"features,omitempty"`
	MFALevel                  GuildMFALevel    `json:"mfa_level"`
	SystemChanneLID           string           `json:"system_channel_id,omitempty"`
	SystemChannelFlags        int              `json:"system_channel_flags"`
	RulesChannelID            string           `json:"rules_channel_id,omitempty"`
	MaxPresences              int              `json:"max_presences,omitempty"`
	MaxMembers                int              `json:"max_members,omitempty"`
	VanityURLCode             string           `json:"vanity_url_code,omitempty"`
	Description               string           `json:"description,omitempty"`
	Banner                    string           `json:"banner,omitempty"`
	PremiumTier               GuildPremiumTier `json:"premium_tier"`
	PremiumSubscriptionCount  int              `json:"premium_subscription_count,omitempty"`
	PreferredLocale           string           `json:"preferred_locale"`
	PublicUpdatesChannelID    string           `json:"public_updates_channel_id,omitempty"`
	MaxVideoChannelUsers      int              `json:"max_video_channel_users,omitempty"`
	MaxStageVideoChannelUsers int              `json:"max_stage_video_channel_users,omitempty"`
	ApproximateMemberCount    int              `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount  int              `json:"approximate_presence_count,omitempty"`
	// WelcomeScreen             *GuildWelcomeScreen     `json:"welcome_screen,omitempty"`
	NSFWLevel GuildNSFWLevel `json:"nsfw_level"`
	// Stickers                  []Sticker               `json:"stickers,omitempty"`
	PremiumProgressBarEnabled bool   `json:"premium_progress_bar_enabled,omitempty"`
	SafetyAlertsChannelID     string `json:"safety_alerts_channel_id,omitempty"`
}

type UnavailableGuild struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable"`
}

type GuildVerificationLevel int

const (
	GuildVerificationLevelNone     GuildVerificationLevel = 0
	GuildVerificationLevelLow      GuildVerificationLevel = 1
	GuildVerificationLevelMedium   GuildVerificationLevel = 2
	GuildVerificationLevelHigh     GuildVerificationLevel = 3
	GuildVerificationLevelVeryHigh GuildVerificationLevel = 4
)

type GuildDefaultNotificationLevel int

const (
	GuildDefaultNotificationLevelAllMessages  GuildDefaultNotificationLevel = 0
	GuildDefaultNotificationLevelOnlyMentions GuildDefaultNotificationLevel = 1
)

type GuildExplicitContentFilterLevel int

const (
	GuildExplicitContentFilterLevelDisabled            GuildExplicitContentFilterLevel = 0
	GuildExplicitContentFilterLevelMembersWithoutRoles GuildExplicitContentFilterLevel = 1
	GuildExplicitContentFilterLevelAllMembers          GuildExplicitContentFilterLevel = 2
)

type GuildMFALevel int

const (
	GuildMFALevelNone     GuildMFALevel = 0
	GuildMFALevelElevated GuildMFALevel = 1
)

type GuildPremiumTier int

const (
	GuildPremiumTierNone GuildPremiumTier = 0
	GuildPremiumTier1    GuildPremiumTier = 1
	GuildPremiumTier2    GuildPremiumTier = 2
	GuildPremiumTier3    GuildPremiumTier = 3
)

type GuildNSFWLevel int

const (
	GuildNSFWLevelDefault       GuildNSFWLevel = 0
	GuildNSFWLevelExplicit      GuildNSFWLevel = 1
	GuildNSFWLevelSafe          GuildNSFWLevel = 2
	GuildNSFWLevelAgeRestricted GuildNSFWLevel = 3
)

type GuildCreateEvent = Guild

type GuildUpdateEvent = Guild

type GuildDeleteEvent = UnavailableGuild

type GuildGetCall struct{}

type GuildGetResponse = Guild

type GuildUpdateCall struct {
	Name                        null.Null[string]                          `json:"name"`
	VerificationLevel           null.Null[GuildVerificationLevel]          `json:"verification_level"`
	DefaultMessageNotifications null.Null[GuildDefaultNotificationLevel]   `json:"default_message_notifications"`
	ExplicitContentFilter       null.Null[GuildExplicitContentFilterLevel] `json:"explicit_content_filter"`
	AFKChannelID                null.Null[string]                          `json:"afk_channel_id"`
	AFKTimeout                  null.Null[int]                             `json:"afk_timeout"`
	Icon                        null.Null[string]                          `json:"icon"`
	Splash                      null.Null[string]                          `json:"splash"`
	DiscoverySplash             null.Null[string]                          `json:"discovery_splash"`
	Banner                      null.Null[string]                          `json:"banner"`
	SystemChannelID             null.Null[string]                          `json:"system_channel_id"`
	SystemChannelFlags          null.Null[int]                             `json:"system_channel_flags"`
	RulesChannelID              null.Null[string]                          `json:"rules_channel_id"`
	PublicUpdatesChannelID      null.Null[string]                          `json:"public_updates_channel_id"`
	PreferredLocale             null.Null[string]                          `json:"preferred_locale"`
	Features                    null.Null[[]string]                        `json:"features"`
	Description                 null.Null[string]                          `json:"description"`
	PremiumProgressBarEnabled   null.Null[bool]                            `json:"premium_progress_bar_enabled"`
	SafetyAlertsChannelID       null.Null[string]                          `json:"safety_alerts_channel_id"`
}

type GuildUpdateResponse = Guild
