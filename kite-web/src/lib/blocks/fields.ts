import { BlockField } from "./types";

// Fields of many Discord blocks, named like the settings of other blocks.

export const guildField: BlockField = {
  name: "guild_target",
  in: "path",
  target: "guild_id",
  type: "snowflake",
  label: "Server",
  description:
    "ID of the server. Leave empty to use the server the flow runs in.",
  fallback: "guild",
};

export const channelField: BlockField = {
  name: "channel_target",
  in: "path",
  target: "channel_id",
  type: "snowflake",
  label: "Channel",
  description: "ID of the channel.",
};

export const flowChannelField: BlockField = {
  ...channelField,
  description:
    "ID of the channel. Leave empty to use the channel the flow runs in.",
  fallback: "channel",
};

export const messageField: BlockField = {
  name: "message_target",
  in: "path",
  target: "message_id",
  type: "snowflake",
  label: "Message",
  description: "ID of the message.",
};

export const userField: BlockField = {
  name: "user_target",
  in: "path",
  target: "user_id",
  type: "snowflake",
  label: "User",
  description: "ID of the user.",
};

export const roleField: BlockField = {
  name: "role_target",
  in: "path",
  target: "role_id",
  type: "snowflake",
  label: "Role",
  description: "ID of the role.",
};
