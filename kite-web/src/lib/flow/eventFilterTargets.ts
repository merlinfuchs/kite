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
  direct_message_create: [
    "message_content",
    "user_id",
    "channel_id",
    "message_id",
    "message_has_attachment",
    "message_is_reply",
    "message_mentions_bot",
  ],
  message_delete: [
    "message_content",
    "user_id",
    "guild_id",
    "channel_id",
    "message_id",
  ],
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
  guild_audit_log_entry_create: [
    "audit_log_action",
    "user_id",
    "guild_id",
    "channel_id",
  ],
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
  audit_log_action: "Audit Log Action",
};

// Yes or no filters, whose value is "true" or "false".
export const booleanEventFilterTargets = [
  "message_has_attachment",
  "message_is_reply",
  "message_mentions_bot",
  "user_is_bot",
];

// getEventFilterModes returns the modes a target can use, undefined if it can
// use all. Yes or no filters say yes or no with their value, so the editor has
// no mode for them, and audit log actions are picked from a list.
export function getEventFilterModes(target?: string) {
  if (!target) return undefined;
  if (booleanEventFilterTargets.includes(target)) return ["equal"];
  if (target === "role_id" || target === "emoji") {
    return ["equal", "not_equal"];
  }
  if (target === "audit_log_action") return ["one_of", "not_one_of"];
  return undefined;
}

// getEventFilterTargets returns the targets available for an event type, or
// every target if the filter isn't connected to an event yet.
export function getEventFilterTargets(eventType?: string) {
  if (!eventType) return Object.keys(eventFilterTargetLabels);
  return eventFilterTargets[eventType] ?? [];
}
