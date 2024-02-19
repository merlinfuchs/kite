package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/merlinfuchs/dismod/distype"

	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

func (b *Bot) registerListeners() {
	b.Cluster.AddEventListener(distype.EventTypeReady, b.handleReady)
	b.Cluster.AddEventListener(distype.EventTypeGuildCreate, b.handleGuildCreate)
	b.Cluster.AddEventListener(distype.EventTypeGuildUpdate, b.handleGuildUpdate)

	b.Cluster.AddAllEventListener(b.handleAny)
}

func (b *Bot) handleReady(s int, e interface{}) {
	slog.Info("Shard is ready", "shard_id", s)
}

func (b *Bot) handleGuildCreate(s int, e interface{}) {
	g := e.(*distype.GuildCreateEvent)

	_, err := b.guildStore.UpsertGuild(context.Background(), model.Guild{
		ID:          g.ID,
		Name:        g.Name,
		Icon:        g.Icon.Value,
		Description: g.Description.Value,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to upsert guild from guild create event")
	}
}

func (b *Bot) handleGuildUpdate(s int, e interface{}) {
	g := e.(*distype.GuildUpdateEvent)

	_, err := b.guildStore.UpsertGuild(context.Background(), model.Guild{
		ID:          g.ID,
		Name:        g.Name,
		Icon:        g.Icon.Value,
		Description: g.Description.Value,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to upsert guild from guild update event")
	}
}

func (b *Bot) handleAny(s int, t distype.EventType, e interface{}) {
	eventType, exists := eventTypeFromEventType(t)
	if !exists {
		return
	}

	guildID := guildIDFromEvent(e)
	if guildID == nil {
		return
	}

	b.Engine.HandleEvent(context.Background(), &event.Event{
		Type:    eventType,
		GuildID: *guildID,
		Data:    e,
	})
}

func guildIDFromEvent(e interface{}) *distype.Snowflake {
	// NOTE: These are all type aliases and some point to the same type, so not all events have to be covered here.

	switch e := e.(type) {
	case *distype.ChannelCreateEvent: // ChannelCreateEvent, ChannelUpdateEvent, ChannelDeleteEvent, ThreadCreateEvent, ThreadUpdateEvent, ThreadDeleteEvent
		return e.GuildID
	case *distype.ChannelPinsUpdateEvent: // ChannelPinsUpdateEvent
		return e.GuildID
	case *distype.ThreadListSyncEvent: // ThreadListSyncEvent
		return &e.GuildID
	case *distype.ThreadMemberUpdateEvent: // ThreadMemberUpdateEvent
		return e.GuildID
	case *distype.ThreadMembersUpdateEvent: // ThreadMembersUpdateEvent
		return &e.GuildID
	case *distype.GuildCreateEvent: // GuildCreateEvent, GuildUpdateEvent
		return &e.ID
	case *distype.GuildDeleteEvent: // GuildDeleteEvent
		return &e.ID
	case *distype.BanAddEvent: // BanAddEvent
		return &e.GuildID
	case *distype.BanRemoveEvent: // BanRemoveEvent
		return &e.GuildID
	case *distype.GuildEmojisUpdateEvent: // EmojisUpdateEvent
		return &e.GuildID
	case *distype.GuildStickersUpdateEvent: // StickersUpdateEvent
		return &e.GuildID
	case *distype.MemberAddEvent: // MemberAddEvent
		return &e.GuildID
	case *distype.MemberUpdateEvent: // MemberUpdateEvent
		return &e.GuildID
	case *distype.MemberRemoveEvent: // MemberRemoveEvent
		return &e.GuildID
	case *distype.RoleCreateEvent: // RoleCreateEvent
		return &e.GuildID
	case *distype.RoleUpdateEvent: // RoleUpdateEvent
		return &e.GuildID
	case *distype.RoleDeleteEvent: // RoleDeleteEvent
		return &e.GuildID
	case *distype.ScheduledEventCreateEvent: // ScheduledEventCreateEvent, ScheduledEventUpdateEvent, ScheduledEventDeleteEvent
		return &e.GuildID
	case *distype.ScheduledEventUserAddEvent: // ScheduledEventUserAddEvent
		return &e.GuildID
	case *distype.ScheduledEventUserRemoveEvent: // ScheduledEventUserRemoveEvent
		return &e.GuildID
	case *distype.InviteCreateEvent: // InviteCreateEvent
		return e.GuildID
	case *distype.InviteDeleteEvent: // InviteDeleteEvent
		return e.GuildID
	case *distype.InteractionCreateEvent: // InteractionCreateEvent
		return e.GuildID
	case *distype.MessageCreateEvent: // MessageCreateEvent, MessageUpdateEvent
		return e.GuildID
	case *distype.MessageDeleteEvent: // MessageDeleteEvent
		return e.GuildID
	case *distype.MessageDeleteBulkEvent: // MessageDeleteBulkEvent
		return e.GuildID
	case *distype.MessageReactionAddEvent: // MessageReactionAddEvent
		return e.GuildID
	case *distype.MessageReactionRemoveEvent: // MessageReactionRemoveEvent
		return e.GuildID
	case *distype.MessageReactionRemoveAllEvent: // MessageReactionRemoveAllEvent
		return e.GuildID
	case *distype.MessageReactionRemoveEmojiEvent: // MessageReactionRemoveEmojiEvent
		return e.GuildID
	case *distype.StageInstanceCreateEvent: // StageInstanceCreateEvent, StageInstanceDeleteEvent, StageInstanceUpdateEvent
		return &e.GuildID
	}

	slog.Warn("Unknown event type to forward to engine", "event", fmt.Sprintf("%T", e))
	return nil
}

func eventTypeFromEventType(t distype.EventType) (event.EventType, bool) {
	switch t {
	case distype.EventTypeChannelCreate:
		return event.DiscordChannelCreate, true
	case distype.EventTypeChannelUpdate:
		return event.DiscordChannelUpdate, true
	case distype.EventTypeChannelDelete:
		return event.DiscordChannelDelete, true
	case distype.EventTypeChannelPinsUpdate:
		return event.DiscordChannelPinsUpdate, true
	case distype.EventTypeThreadCreate:
		return event.DiscordThreadCreate, true
	case distype.EventTypeThreadUpdate:
		return event.DiscordThreadUpdate, true
	case distype.EventTypeThreadDelete:
		return event.DiscordThreadDelete, true
	case distype.EventTypeThreadListSync:
		return event.DiscordThreadListSync, true
	case distype.EventTypeThreadMemberUpdate:
		return event.DiscordThreadMemberUpdate, true
	case distype.EventTypeThreadMembersUpdate:
		return event.DiscordThreadMembersUpdate, true
	case distype.EventTypeGuildCreate:
		return event.DiscordGuildCreate, true
	case distype.EventTypeGuildUpdate:
		return event.DiscordGuildUpdate, true
	case distype.EventTypeGuildDelete:
		return event.DiscordGuildDelete, true
	case distype.EventTypeGuildBanAdd:
		return event.DiscordGuildBanAdd, true
	case distype.EventTypeGuildBanRemove:
		return event.DiscordGuildBanRemove, true
	case distype.EventTypeGuildEmojisUpdate:
		return event.DiscordGuildEmojisUpdate, true
	case distype.EventTypeGuildStickersUpdate:
		return event.DiscordGuildStickersUpdate, true
	case distype.EventTypeGuildMemberAdd:
		return event.DiscordGuildMemberAdd, true
	case distype.EventTypeGuildMemberRemove:
		return event.DiscordGuildMemberRemove, true
	case distype.EventTypeGuildMemberUpdate:
		return event.DiscordGuildMemberUpdate, true
	case distype.EventTypeGuildRoleCreate:
		return event.DiscordGuildRoleCreate, true
	case distype.EventTypeGuildRoleUpdate:
		return event.DiscordGuildRoleUpdate, true
	case distype.EventTypeGuildRoleDelete:
		return event.DiscordGuildRoleDelete, true
	case distype.EventTypeGuildScheduledEventCreate:
		return event.DiscordGuildScheduledEventCreate, true
	case distype.EventTypeGuildScheduledEventUpdate:
		return event.DiscordGuildScheduledEventUpdate, true
	case distype.EventTypeGuildScheduledEventDelete:
		return event.DiscordGuildScheduledEventDelete, true
	case distype.EventTypeGuildScheduledEventUserAdd:
		return event.DiscordGuildScheduledEventUserAdd, true
	case distype.EventTypeGuildScheduledEventUserRemove:
		return event.DiscordGuildScheduledEventUserRemove, true
	case distype.EventTypeInviteCreate:
		return event.DiscordInviteCreate, true
	case distype.EventTypeInviteDelete:
		return event.DiscordInviteDelete, true
	case distype.EventTypeInteractionCreate:
		return event.DiscordInteractionCreate, true
	case distype.EventTypeMessageCreate:
		return event.DiscordMessageCreate, true
	case distype.EventTypeMessageUpdate:
		return event.DiscordMessageUpdate, true
	case distype.EventTypeMessageDelete:
		return event.DiscordMessageDelete, true
	case distype.EventTypeMessageDeleteBulk:
		return event.DiscordMessageDeleteBulk, true
	case distype.EventTypeMessageReactionAdd:
		return event.DiscordMessageReactionAdd, true
	case distype.EventTypeMessageReactionRemove:
		return event.DiscordMessageReactionRemove, true
	case distype.EventTypeMessageReactionRemoveAll:
		return event.DiscordMessageReactionRemoveAll, true
	case distype.EventTypeMessageReactionRemoveEmoji:
		return event.DiscordMessageReactionRemoveEmoji, true
	case distype.EventTypeStageInstanceCreate:
		return event.DiscordStageInstanceCreate, true
	case distype.EventTypeStageInstanceDelete:
		return event.DiscordStageInstanceDelete, true
	case distype.EventTypeStageInstanceUpdate:
		return event.DiscordStageInstanceUpdate, true
	}

	return "", false
}
