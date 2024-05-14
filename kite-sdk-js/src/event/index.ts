import { addEventHandler } from "../sys";
import * as discord from "../discord";

export * from "./types.generated";

type eventTypes = {
  INSTANTIATE: void;
  DISCORD_CHANNEL_CREATE: discord.ChannelCreateEvent;
  DISCORD_CHANNEL_UPDATE: discord.ChannelUpdateEvent;
  DISCORD_CHANNEL_DELETE: discord.ChannelDeleteEvent;
  DISCORD_CHANNEL_PINS_UPDATE: discord.ChannelPinsUpdateEvent;
  DISCORD_THREAD_CREATE: discord.ThreadCreateEvent;
  DISCORD_THREAD_UPDATE: discord.ThreadUpdateEvent;
  DISCORD_THREAD_DELETE: discord.ThreadDeleteEvent;
  DISCORD_THREAD_LIST_SYNC: discord.ThreadListSyncEvent;
  DISCORD_THREAD_MEMBER_UPDATE: discord.ThreadMemberUpdateEvent;
  DISCORD_THREAD_MEMBERS_UPDATE: discord.ThreadMembersUpdateEvent;
  DISCORD_GUILD_CREATE: discord.GuildCreateEvent;
  DISCORD_GUILD_UPDATE: discord.GuildUpdateEvent;
  DISCORD_GUILD_DELETE: discord.GuildDeleteEvent;
  DISCORD_GUILD_BAN_ADD: discord.BanAddEvent;
  DISCORD_GUILD_BAN_REMOVE: discord.BanRemoveEvent;
  DISCORD_GUILD_EMOJIS_UPDATE: discord.GuildEmojisUpdateEvent;
  DISCORD_GUILD_STICKERS_UPDATE: discord.GuildStickersUpdateEvent;
  DISCORD_GUILD_MEMBER_ADD: discord.MemberAddEvent;
  DISCORD_GUILD_MEMBER_REMOVE: discord.MemberRemoveEvent;
  DISCORD_GUILD_MEMBER_UPDATE: discord.MemberUpdateEvent;
  DISCORD_GUILD_ROLE_CREATE: discord.RoleCreateEvent;
  DISCORD_GUILD_ROLE_UPDATE: discord.RoleUpdateEvent;
  DISCORD_GUILD_ROLE_DELETE: discord.RoleDeleteEvent;
  DISCORD_GUILD_SCHEDULED_EVENT_CREATE: discord.ScheduledEventCreateEvent;
  DISCORD_GUILD_SCHEDULED_EVENT_UPDATE: discord.ScheduledEventUpdateEvent;
  DISCORD_GUILD_SCHEDULED_EVENT_DELETE: discord.ScheduledEventDeleteEvent;
  DISCORD_GUILD_SCHEDULED_EVENT_USER_ADD: discord.ScheduledEventUserAddEvent;
  DISCORD_GUILD_SCHEDULED_EVENT_USER_REMOVE: discord.ScheduledEventUserRemoveEvent;
  DISCORD_INVITE_CREATE: discord.InviteCreateEvent;
  DISCORD_INVITE_DELETE: discord.InviteDeleteEvent;
  DISCORD_INTERACTION_CREATE: discord.InteractionCreateEvent;
  DISCORD_MESSAGE_CREATE: discord.MessageCreateEvent;
  DISCORD_MESSAGE_UPDATE: discord.MessageUpdateEvent;
  DISCORD_MESSAGE_DELETE: discord.MessageDeleteEvent;
  DISCORD_MESSAGE_DELETE_BULK: discord.MessageDeleteBulkEvent;
  DISCORD_MESSAGE_REACTION_ADD: discord.MessageReactionAddEvent;
  DISCORD_MESSAGE_REACTION_REMOVE: discord.MessageReactionRemoveEvent;
  DISCORD_MESSAGE_REACTION_REMOVE_ALL: discord.MessageReactionRemoveAllEvent;
  DISCORD_MESSAGE_REACTION_REMOVE_EMOJI: discord.MessageReactionRemoveEmojiEvent;
  DISCORD_STAGE_INSTANCE_CREATE: discord.StageInstanceCreateEvent;
  DISCORD_STAGE_INSTANCE_UPDATE: discord.StageInstanceUpdateEvent;
  DISCORD_STAGE_INSTANCE_DELETE: discord.StageInstanceDeleteEvent;
};

export type EventType = keyof eventTypes;

export function on<T extends EventType>(
  event: T,
  handler: (e: eventTypes[T]) => void
) {
  addEventHandler(event, handler);
}
