package dismodel

import "time"

type Invite struct {
	Code                       string           `json:"code"`
	Guild                      *Guild           `json:"guild"`
	Channel                    *Channel         `json:"channel"`
	Inviter                    *User            `json:"inviter"`
	TargetType                 InviteTargetType `json:"target_type"`
	TargetUser                 *User            `json:"target_user"`
	TargetApp                  *Application     `json:"target_application"`
	ApproximatePresenceCounter int              `json:"approximate_presence_count"`
	ApproximateMemberCounter   int              `json:"approximate_member_count"`
	ExpiresAt                  time.Time        `json:"expires_at"`
	// StageInstance              *StageInstance   `json:"stage_instance"`
	// GuildScheduledEvent *GuildScheduledEvent `json:"guild_scheduled_event"`
}

type InviteTargetType int

const (
	InviteTargetTypeStream              InviteTargetType = 1
	InviteTargetTypeEmbeddedApplication InviteTargetType = 2
)

type InviteCreateEvent struct {
	ChannelID         string           `json:"channel_id"`
	Code              string           `json:"code"`
	CreatedAt         time.Time        `json:"created_at"`
	GuildID           string           `json:"guild_id"`
	Inviter           *User            `json:"inviter"`
	MaxAge            int              `json:"max_age"`
	MaxUses           int              `json:"max_uses"`
	TargetType        InviteTargetType `json:"target_type"`
	TargetUser        *User            `json:"target_user"`
	TargetApplication *Application     `json:"target_application"`
	Temporary         bool             `json:"temporary"`
	Uses              int              `json:"uses"`
}

type InviteDeleteEvent struct {
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id"`
	Code      string `json:"code"`
}

type InviteListForChannelCall struct {
	ChannelID string `json:"channel_id"`
}

type InviteListForChannelResponse = []Invite

type InviteListForGuildCall struct{}

type InviteListForGuildResponse = []Invite

type InviteCreateCall struct {
	ChannelID           string           `json:"channel_id"`
	MaxAge              int              `json:"max_age,omitempty"`
	MaxUses             int              `json:"max_uses,omitempty"`
	Temporary           bool             `json:"temporary,omitempty"`
	Unique              bool             `json:"unique,omitempty"`
	TargetType          InviteTargetType `json:"target_type,omitempty"`
	TargetUserID        string           `json:"target_user_id,omitempty"`
	TargetApplicationID string           `json:"target_application_id,omitempty"`
}

type InviteCreateResponse = Invite

type InviteGetCall struct {
	Code                  string `json:"code"`
	WithCounts            bool   `json:"with_counts,omitempty"`
	WithExpiration        bool   `json:"with_expiration,omitempty"`
	GuildScheduledEventID string `json:"guild_scheduled_event_id,omitempty"`
}

type InviteGetResponse = Invite

type InviteDeleteCall struct {
	Code string `json:"code"`
}

type InviteDeleteResponse = Invite
