// What event filters can check for each event type, matching eventFilterTargets
// in the service. A target that isn't listed doesn't exist for the event, and
// the service rejects it on save.
export const eventFilterTargets: Record<string, string[]> = {
  message_create: [
    "message_content",
    "user_id",
    "guild_id",
    "channel_id",
    "message_id",
    "role_id",
    "message_has_attachment",
    "message_is_reply",
    "message_mentions_bot",
  ],
  message_update: [
    "message_content",
    "user_id",
    "guild_id",
    "channel_id",
    "message_id",
    "role_id",
    "message_has_attachment",
    "message_is_reply",
    "message_mentions_bot",
  ],
  message_delete: ["guild_id", "channel_id", "message_id"],
  guild_member_add: ["user_id", "guild_id", "role_id", "user_is_bot"],
  guild_member_remove: ["user_id", "guild_id", "user_is_bot"],
  message_reaction_add: [
    "user_id",
    "guild_id",
    "channel_id",
    "message_id",
    "role_id",
    "emoji",
    "user_is_bot",
  ],
  message_reaction_remove: [
    "user_id",
    "guild_id",
    "channel_id",
    "message_id",
    "emoji",
  ],
  guild_create: ["guild_id"],
  guild_delete: ["guild_id"],
  guild_boost: ["user_id", "guild_id", "channel_id", "message_id", "role_id"],
};

export const eventFilterTargetLabels: Record<string, string> = {
  message_content: "Message Content",
  user_id: "User ID",
  guild_id: "Guild ID",
  channel_id: "Channel ID",
  message_id: "Message ID",
  role_id: "Member Role",
  emoji: "Reaction Emoji",
  message_has_attachment: "Message Has Attachment",
  message_is_reply: "Message Is Reply",
  message_mentions_bot: "Message Mentions Bot",
  user_is_bot: "User Is Bot",
};

// Yes or no filters, whose value is "true" or "false".
export const booleanEventFilterTargets = [
  "message_has_attachment",
  "message_is_reply",
  "message_mentions_bot",
  "user_is_bot",
];

// Filters that can only be equal or not equal.
export const equalityEventFilterTargets = [
  "role_id",
  "emoji",
  ...booleanEventFilterTargets,
];

// getEventFilterTargets returns the targets available for an event type, or
// every target if the filter isn't connected to an event yet.
export function getEventFilterTargets(eventType?: string) {
  if (!eventType) return Object.keys(eventFilterTargetLabels);
  return eventFilterTargets[eventType] ?? [];
}
