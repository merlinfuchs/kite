import { AnyZodObject, z } from "zod";
import {
  channelTargetSchema,
  guildTargetSchema,
  messageDataSchema,
  messageTargetSchema,
  roleTargetSchema,
  templated,
  userPicked,
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

// Settings of custom blocks, with the schema and editor input the setting
// has in every block.

export const guildTargetSetting: BlockField = {
  name: "guild_target",
  type: "snowflake",
  schema: guildTargetSchema.optional(),
};

export const channelTargetSetting: BlockField = {
  name: "channel_target",
  type: "snowflake",
  schema: channelTargetSchema,
};

export const userTargetSetting: BlockField = {
  name: "user_target",
  type: "snowflake",
  schema: userTargetSchema,
};

export const roleTargetSetting: BlockField = {
  name: "role_target",
  type: "snowflake",
  schema: roleTargetSchema,
};

// The fields of requests with the same schema and editor input, for blocks
// that were written by hand before they were requests.

export const guildTargetField: BlockField = {
  ...guildField,
  schema: guildTargetSetting.schema,
};

export const channelTargetField: BlockField = {
  ...channelField,
  schema: channelTargetSetting.schema,
};

export const messageTargetField: BlockField = {
  ...messageField,
  schema: messageTargetSchema,
};

export const userTargetField: BlockField = {
  ...userField,
  schema: userTargetSetting.schema,
};

export const roleTargetField: BlockField = {
  ...roleField,
  schema: roleTargetSetting.schema,
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

// The settings of a condition's branch, edited by the inputs of its kind,
// like condition_item_user_mode.
export function conditionItemFields(
  kind: string,
  modeSchema: z.ZodTypeAny,
  valueDescription: string
): BlockField[] {
  return [
    {
      name: "condition_item_mode",
      type: "string",
      input: `condition_item_${kind}_mode`,
      schema: modeSchema,
    },
    {
      name: "condition_item_value",
      type: "string",
      input: `condition_item_${kind}_value`,
      schema: templated(z.string(), valueDescription).optional(),
    },
  ];
}

// Message blocks send either an inline message or a saved template, see
// requireMessage.
export const messageDataField: BlockField = {
  name: "message_data",
  type: "string",
  schema: messageDataSchema.optional(),
};

export const messageTemplateField: BlockField = {
  name: "message_template_id",
  type: "string",
  schema: userPicked(
    z.string(),
    "ID of a saved message template to send instead of message_data."
  ).optional(),
};

export function requireMessage(schema: AnyZodObject) {
  return schema
    .refine(
      (data) => !!data.message_data || !!data.message_template_id,
      "Either message_data or message_template_id is required"
    )
    .describe("Set either message_data or message_template_id.");
}
