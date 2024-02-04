package dismodel

import (
	"time"

	"github.com/merlinfuchs/kite/kite-types/null"
)

type ScheduledEvent struct {
	ID                 string                        `json:"id"`
	GuildID            string                        `json:"guild_id"`
	ChannelID          string                        `json:"channel_id"`
	CreatorID          string                        `json:"creator_id"`
	Name               string                        `json:"name"`
	Description        string                        `json:"description"`
	ScheduledStartTime time.Time                     `json:"scheduled_start_time"`
	ScheduledEndTime   time.Time                     `json:"scheduled_end_time"`
	PrivacyLevel       ScheduledEventPrivacyLevel    `json:"privacy_level"`
	Status             ScheduledEventStatus          `json:"status"`
	EntityType         ScheduledEventEntityType      `json:"entity_type"`
	EntityID           string                        `json:"entity_id"`
	EntityMetadata     *ScheduledEventEntityMetadata `json:"entity_metadata"`
	Creator            *User                         `json:"creator"`
	UserCount          int                           `json:"user_count"`
	Image              string                        `json:"image"`
}

type ScheduledEventPrivacyLevel int

const (
	ScheduledEventPrivacyLevelGuildOnly ScheduledEventPrivacyLevel = 2
)

type ScheduledEventStatus int

const (
	ScheduledEventStatusScheduled ScheduledEventStatus = 1
	ScheduledEventStatusActive    ScheduledEventStatus = 2
	ScheduledEventStatusCompleted ScheduledEventStatus = 3
	ScheduledEventStatusCanceled  ScheduledEventStatus = 4
)

type ScheduledEventEntityType int

const (
	ScheduledEventEntityTypeStageInstance ScheduledEventEntityType = 1
	ScheduledEventEntityTypeVoice         ScheduledEventEntityType = 2
	ScheduledEventEntityTypeExternal      ScheduledEventEntityType = 3
)

type ScheduledEventEntityMetadata struct {
	Location string `json:"location"`
}

type ScheduledEventUser struct {
	GuildScheduledEventID string  `json:"guild_scheduled_event_id"`
	User                  *User   `json:"user"`
	Member                *Member `json:"member,omitempty"`
}

type GuildScheduledEventCreateEvent = ScheduledEvent

type GuildScheduledEventUpdateEvent = ScheduledEvent

type GuildScheduledEventDeleteEvent = ScheduledEvent

type GuildScheduledEventStatusUpdateEvent = ScheduledEvent

type GuildScheduledEventUserAddEvent struct {
	GuildScheduledEventID string `json:"guild_scheduled_event_id"`
	UserID                string `json:"user_id"`
	GuildID               string `json:"guild_id"`
}

type GuildScheduledEventUserRemoveEvent struct {
	GuildScheduledEventID string `json:"guild_scheduled_event_id"`
	UserID                string `json:"user_id"`
	GuildID               string `json:"guild_id"`
}

type ScheduledEventListCall struct {
	WithUserCount bool `json:"with_user_count,omitempty"`
}

type ScheduledEventListResponse = []ScheduledEvent

type ScheduledEventCreateCall struct {
	ChannelID          string                        `json:"channel_id,omitempty"`
	EntityMetdata      *ScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	Name               string                        `json:"name"`
	PrivacyLevel       ScheduledEventPrivacyLevel    `json:"privacy_level"`
	ScheduledStartTime time.Time                     `json:"scheduled_start_time"`
	ScheduledEndTime   time.Time                     `json:"scheduled_end_time,omitempty"`
	Description        string                        `json:"description,omitempty"`
	EntityType         ScheduledEventEntityType      `json:"entity_type,omitempty"`
	Image              string                        `json:"image,omitempty"`
}

type ScheduledEventCreateResponse = ScheduledEvent

type ScheduledEventGetCall struct {
	ID string `json:"id"`
}

type ScheduledEventGetResponse = ScheduledEvent

type ScheduledEventUpdateCall struct {
	ID                 string                                `json:"id"`
	ChannelID          null.Null[string]                     `json:"channel_id,omitempty"`
	Name               null.Null[string]                     `json:"name,omitempty"`
	PrivacyLevel       null.Null[ScheduledEventPrivacyLevel] `json:"privacy_level,omitempty"`
	ScheduledStartTime null.Null[time.Time]                  `json:"scheduled_start_time,omitempty"`
	ScheduledEndTime   null.Null[time.Time]                  `json:"scheduled_end_time,omitempty"`
	Description        null.Null[string]                     `json:"description,omitempty"`
	EntityType         null.Null[ScheduledEventEntityType]   `json:"entity_type,omitempty"`
	Status             null.Null[ScheduledEventStatus]       `json:"status,omitempty"`
	Image              null.Null[string]                     `json:"image,omitempty"`
}

type ScheduledEventUpdateResponse = ScheduledEvent

type ScheduledEventDeleteCall struct {
	ID string `json:"id"`
}

type ScheduledEventDeleteResponse struct{}

type ScheduledEventUserListCall struct {
	ID         string `json:"id"`
	Limit      int    `json:"limit,omitempty"`
	WithMember bool   `json:"with_member,omitempty"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
}

type ScheduledEventUserListResponse = []ScheduledEventUser
