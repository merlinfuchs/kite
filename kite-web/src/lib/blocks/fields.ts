import { z } from "zod";
import {
  channelTargetSchema,
  guildTargetSchema,
  messageTargetSchema,
  roleTargetSchema,
  templated,
  userTargetSchema,
} from "../flow/dataSchema";
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

// The fields above with the schema and editor input the setting has in blocks
// that aren't generated from their fields.

export const guildTargetField: BlockField = {
  ...guildField,
  schema: guildTargetSchema.optional(),
};

export const channelTargetField: BlockField = {
  ...channelField,
  schema: channelTargetSchema,
};

export const messageTargetField: BlockField = {
  ...messageField,
  schema: messageTargetSchema,
};

export const userTargetField: BlockField = {
  ...userField,
  schema: userTargetSchema,
};

export const roleTargetField: BlockField = {
  ...roleField,
  schema: roleTargetSchema,
};

// The settings of a condition, edited by the inputs of its kind, like
// condition_user_base_value.
export function conditionFields(
  kind: string,
  baseValueDescription: string
): BlockField[] {
  return [
    {
      name: "condition_base_value",
      type: "string",
      input: `condition_${kind}_base_value`,
      schema: templated(z.string(), baseValueDescription),
    },
    {
      name: "condition_allow_multiple",
      type: "boolean",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether every matching branch runs. If unset, only the first matching branch runs."
        ),
    },
  ];
}
