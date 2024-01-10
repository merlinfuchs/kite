package dismodel

import "github.com/merlinfuchs/kite/go-types/null"

type StageInstance struct {
	ID                    string            `json:"id"`
	GuildID               string            `json:"guild_id"`
	ChannelID             string            `json:"channel_id"`
	Topic                 string            `json:"topic"`
	PrivacyLevel          StagePrivacyLevel `json:"privacy_level"`
	DiscoverableDisabled  bool              `json:"discoverable_disabled"`
	GuildScheduledEventID string            `json:"guild_scheduled_event_id"`
}

type StagePrivacyLevel int

const (
	StagePrivacyLevelPublic    StagePrivacyLevel = 1
	StagePrivacyLevelGuildOnly StagePrivacyLevel = 2
)

type StageInstanceCreateEvent = StageInstance

type StageInstanceUpdateEvent = StageInstance

type StageInstanceDeleteEvent = StageInstance

type StageInstanceCreateCall struct {
	ChannelID             string            `json:"channel_id"`
	Topic                 string            `json:"topic"`
	PrivacyLevel          StagePrivacyLevel `json:"privacy_level"`
	SendStartNotification bool              `json:"send_start_notification"`
	GuildSCheduledEventID string            `json:"guild_scheduled_event_id"`
}

type StageInstanceCreateResponse = StageInstance

type StageInstanceGetCall struct {
	ChannelID string `json:"channel_id"`
}

type StageInstanceGetResponse = StageInstance

type StageInstanceUpdateCall struct {
	ChannelID    string                       `json:"channel_id"`
	Topic        null.Null[string]            `json:"topic,omitempty"`
	PrivacyLevel null.Null[StagePrivacyLevel] `json:"privacy_level,omitempty"`
}

type StageInstanceUpdateResponse = StageInstance

type StageInstanceDeleteCall struct {
	ChannelID string `json:"channel_id"`
}

type StageInstanceDeleteResponse struct{}
