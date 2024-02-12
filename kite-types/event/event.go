package event

import (
	"encoding/json"

	"github.com/merlinfuchs/dismod/distype"

	"github.com/merlinfuchs/kite/kite-types/fail"
	"github.com/merlinfuchs/kite/kite-types/internal"
)

type EventHandler func(event Event) error

type EventType string

const (
	Instantiate                          EventType = "INSTANTIATE"
	DiscordChannelCreate                 EventType = "DISCORD_CHANNEL_CREATE"
	DiscordChannelUpdate                 EventType = "DISCORD_CHANNEL_UPDATE"
	DiscordChannelDelete                 EventType = "DISCORD_CHANNEL_DELETE"
	DiscordChannelPinsUpdate             EventType = "DISCORD_CHANNEL_PINS_UPDATE"
	DiscordThreadCreate                  EventType = "DISCORD_THREAD_CREATE"
	DiscordThreadUpdate                  EventType = "DISCORD_THREAD_UPDATE"
	DiscordThreadDelete                  EventType = "DISCORD_THREAD_DELETE"
	DiscordThreadListSync                EventType = "DISCORD_THREAD_LIST_SYNC"
	DiscordThreadMemberUpdate            EventType = "DISCORD_THREAD_MEMBER_UPDATE"
	DiscordThreadMembersUpdate           EventType = "DISCORD_THREAD_MEMBERS_UPDATE"
	DiscordGuildCreate                   EventType = "DISCORD_GUILD_CREATE"
	DiscordGuildUpdate                   EventType = "DISCORD_GUILD_UPDATE"
	DiscordGuildDelete                   EventType = "DISCORD_GUILD_DELETE"
	DiscordGuildBanAdd                   EventType = "DISCORD_GUILD_BAN_ADD"
	DiscordGuildBanRemove                EventType = "DISCORD_GUILD_BAN_REMOVE"
	DiscordGuildEmojisUpdate             EventType = "DISCORD_GUILD_EMOJIS_UPDATE"
	DiscordGuildStickersUpdate           EventType = "DISCORD_GUILD_STICKERS_UPDATE"
	DiscordGuildMemberAdd                EventType = "DISCORD_GUILD_MEMBER_ADD"
	DiscordGuildMemberRemove             EventType = "DISCORD_GUILD_MEMBER_REMOVE"
	DiscordGuildMemberUpdate             EventType = "DISCORD_GUILD_MEMBER_UPDATE"
	DiscordGuildRoleCreate               EventType = "DISCORD_GUILD_ROLE_CREATE"
	DiscordGuildRoleUpdate               EventType = "DISCORD_GUILD_ROLE_UPDATE"
	DiscordGuildRoleDelete               EventType = "DISCORD_GUILD_ROLE_DELETE"
	DiscordGuildScheduledEventCreate     EventType = "DISCORD_GUILD_SCHEDULED_EVENT_CREATE"
	DiscordGuildScheduledEventUpdate     EventType = "DISCORD_GUILD_SCHEDULED_EVENT_UPDATE"
	DiscordGuildScheduledEventDelete     EventType = "DISCORD_GUILD_SCHEDULED_EVENT_DELETE"
	DiscordGuildScheduledEventUserAdd    EventType = "DISCORD_GUILD_SCHEDULED_EVENT_USER_ADD"
	DiscordGuildScheduledEventUserRemove EventType = "DISCORD_GUILD_SCHEDULED_EVENT_USER_REMOVE"
	DiscordInviteCreate                  EventType = "DISCORD_INVITE_CREATE"
	DiscordInviteDelete                  EventType = "DISCORD_INVITE_DELETE"
	DiscordInteractionCreate             EventType = "DISCORD_INTERACTION_CREATE"
	DiscordMessageCreate                 EventType = "DISCORD_MESSAGE_CREATE"
	DiscordMessageUpdate                 EventType = "DISCORD_MESSAGE_UPDATE"
	DiscordMessageDelete                 EventType = "DISCORD_MESSAGE_DELETE"
	DiscordMessageDeleteBulk             EventType = "DISCORD_MESSAGE_DELETE_BULK"
	DiscordMessageReactionAdd            EventType = "DISCORD_MESSAGE_REACTION_ADD"
	DiscordMessageReactionRemove         EventType = "DISCORD_MESSAGE_REACTION_REMOVE"
	DiscordMessageReactionRemoveAll      EventType = "DISCORD_MESSAGE_REACTION_REMOVE_ALL"
	DiscordMessageReactionRemoveEmoji    EventType = "DISCORD_MESSAGE_REACTION_REMOVE_EMOJI"
	DiscordStageInstanceCreate           EventType = "DISCORD_STAGE_INSTANCE_CREATE"
	DiscordStageInstanceUpdate           EventType = "DISCORD_STAGE_INSTANCE_UPDATE"
	DiscordStageInstanceDelete           EventType = "DISCORD_STAGE_INSTANCE_DELETE"
)

type Event struct {
	Type    EventType         `json:"type"`
	GuildID distype.Snowflake `json:"guild_id"`
	Data    interface{}       `json:"data"`
}

func (e *Event) UnmarshalJSON(b []byte) error {
	var temp struct {
		Type EventType       `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	e.Type = temp.Type

	var err error

	switch e.Type {
	case DiscordChannelCreate:
		e.Data, err = internal.DecodeT[distype.ChannelCreateEvent](temp.Data)
	case DiscordChannelUpdate:
		e.Data, err = internal.DecodeT[distype.ChannelUpdateEvent](temp.Data)
	case DiscordChannelDelete:
		e.Data, err = internal.DecodeT[distype.ChannelDeleteEvent](temp.Data)
	case DiscordChannelPinsUpdate:
		e.Data, err = internal.DecodeT[distype.ChannelPinsUpdateEvent](temp.Data)
	case DiscordThreadCreate:
		e.Data, err = internal.DecodeT[distype.ThreadCreateEvent](temp.Data)
	case DiscordThreadUpdate:
		e.Data, err = internal.DecodeT[distype.ThreadUpdateEvent](temp.Data)
	case DiscordThreadDelete:
		e.Data, err = internal.DecodeT[distype.ThreadDeleteEvent](temp.Data)
	case DiscordThreadListSync:
		e.Data, err = internal.DecodeT[distype.ThreadListSyncEvent](temp.Data)
	case DiscordThreadMemberUpdate:
		e.Data, err = internal.DecodeT[distype.ThreadMemberUpdateEvent](temp.Data)
	case DiscordThreadMembersUpdate:
		e.Data, err = internal.DecodeT[distype.ThreadMembersUpdateEvent](temp.Data)
	case DiscordGuildCreate:
		e.Data, err = internal.DecodeT[distype.GuildCreateEvent](temp.Data)
	case DiscordGuildUpdate:
		e.Data, err = internal.DecodeT[distype.GuildUpdateEvent](temp.Data)
	case DiscordGuildDelete:
		e.Data, err = internal.DecodeT[distype.GuildDeleteEvent](temp.Data)
	case DiscordGuildBanAdd:
		e.Data, err = internal.DecodeT[distype.BanAddEvent](temp.Data)
	case DiscordGuildBanRemove:
		e.Data, err = internal.DecodeT[distype.BanRemoveEvent](temp.Data)
	case DiscordGuildEmojisUpdate:
		e.Data, err = internal.DecodeT[distype.GuildEmojisUpdateEvent](temp.Data)
	case DiscordGuildStickersUpdate:
		e.Data, err = internal.DecodeT[distype.GuildStickersUpdateEvent](temp.Data)
	case DiscordGuildMemberAdd:
		e.Data, err = internal.DecodeT[distype.MemberAddEvent](temp.Data)
	case DiscordGuildMemberRemove:
		e.Data, err = internal.DecodeT[distype.MemberRemoveEvent](temp.Data)
	case DiscordGuildMemberUpdate:
		e.Data, err = internal.DecodeT[distype.MemberUpdateEvent](temp.Data)
	case DiscordGuildRoleCreate:
		e.Data, err = internal.DecodeT[distype.RoleCreateEvent](temp.Data)
	case DiscordGuildRoleUpdate:
		e.Data, err = internal.DecodeT[distype.RoleUpdateEvent](temp.Data)
	case DiscordGuildRoleDelete:
		e.Data, err = internal.DecodeT[distype.RoleDeleteEvent](temp.Data)
	case DiscordGuildScheduledEventCreate:
		e.Data, err = internal.DecodeT[distype.ScheduledEventCreateEvent](temp.Data)
	case DiscordGuildScheduledEventUpdate:
		e.Data, err = internal.DecodeT[distype.ScheduledEventUpdateEvent](temp.Data)
	case DiscordGuildScheduledEventDelete:
		e.Data, err = internal.DecodeT[distype.ScheduledEventDeleteEvent](temp.Data)
	case DiscordGuildScheduledEventUserAdd:
		e.Data, err = internal.DecodeT[distype.ScheduledEventUserAddEvent](temp.Data)
	case DiscordGuildScheduledEventUserRemove:
		e.Data, err = internal.DecodeT[distype.ScheduledEventUserRemoveEvent](temp.Data)
	case DiscordInviteCreate:
		e.Data, err = internal.DecodeT[distype.InviteCreateEvent](temp.Data)
	case DiscordInviteDelete:
		e.Data, err = internal.DecodeT[distype.InviteDeleteEvent](temp.Data)
	case DiscordInteractionCreate:
		e.Data, err = internal.DecodeT[distype.InteractionCreateEvent](temp.Data)
	case DiscordMessageCreate:
		e.Data, err = internal.DecodeT[distype.MessageCreateEvent](temp.Data)
	case DiscordMessageUpdate:
		e.Data, err = internal.DecodeT[distype.MessageUpdateEvent](temp.Data)
	case DiscordMessageDelete:
		e.Data, err = internal.DecodeT[distype.MessageDeleteEvent](temp.Data)
	case DiscordMessageDeleteBulk:
		e.Data, err = internal.DecodeT[distype.MessageDeleteBulkEvent](temp.Data)
	case DiscordMessageReactionAdd:
		e.Data, err = internal.DecodeT[distype.MessageReactionAddEvent](temp.Data)
	case DiscordMessageReactionRemove:
		e.Data, err = internal.DecodeT[distype.MessageReactionRemoveEvent](temp.Data)
	case DiscordMessageReactionRemoveAll:
		e.Data, err = internal.DecodeT[distype.MessageReactionRemoveAllEvent](temp.Data)
	case DiscordMessageReactionRemoveEmoji:
		e.Data, err = internal.DecodeT[distype.MessageReactionRemoveEmojiEvent](temp.Data)
	case DiscordStageInstanceCreate:
		e.Data, err = internal.DecodeT[distype.StageInstanceCreateEvent](temp.Data)
	case DiscordStageInstanceUpdate:
		e.Data, err = internal.DecodeT[distype.StageInstanceUpdateEvent](temp.Data)
	case DiscordStageInstanceDelete:
		e.Data, err = internal.DecodeT[distype.StageInstanceDeleteEvent](temp.Data)
	}

	return err
}

type EventResponse struct {
	Success bool              `json:"success"`
	Error   *fail.ModuleError `json:"error"`
}

func EventError(err error) EventResponse {
	return EventResponse{
		Success: false,
		Error:   &fail.ModuleError{Message: err.Error()},
	}
}
