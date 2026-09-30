import { AnyZodObject, z } from "zod";
import {
  channelDataSchema,
  channelTargetSchema,
  guildTargetSchema,
  messageDataSchema,
  messageTargetSchema,
  responseTargetSchema,
  roleTargetSchema,
  templated,
  userPicked,
  userTargetSchema,
  variableIdSchema,
  variableScopeSchema,
} from "../flow/dataSchema";
import { BlockField } from "./types";

// Settings many blocks share, with the schema and editor input they have in
// every block.

export const guildTargetSetting: BlockField = {
  name: "guild_target",
  schema: guildTargetSchema.optional(),
};

export const channelTargetSetting: BlockField = {
  name: "channel_target",
  schema: channelTargetSchema,
};

export const messageTargetSetting: BlockField = {
  name: "message_target",
  schema: messageTargetSchema,
};

export const userTargetSetting: BlockField = {
  name: "user_target",
  schema: userTargetSchema,
};

export const roleTargetSetting: BlockField = {
  name: "role_target",
  schema: roleTargetSchema,
};

// The response message, edited by its own input.
export const responseTargetSetting: BlockField = {
  name: "message_target",
  input: "response_target",
  schema: responseTargetSchema,
};

export const channelDataSetting: BlockField = {
  name: "channel_data",
  schema: channelDataSchema,
};

export const variableSettings: BlockField[] = [
  { name: "variable_id", schema: variableIdSchema },
  { name: "variable_scope", schema: variableScopeSchema },
];

// The shared settings as path parameters of requests.

export const guildTargetField: BlockField = {
  ...guildTargetSetting,
  in: "path",
  target: "guild_id",
  type: "snowflake",
  fallback: "guild",
};

export const channelTargetField: BlockField = {
  ...channelTargetSetting,
  in: "path",
  target: "channel_id",
  type: "snowflake",
};

export const messageTargetField: BlockField = {
  ...messageTargetSetting,
  in: "path",
  target: "message_id",
  type: "snowflake",
};

export const userTargetField: BlockField = {
  ...userTargetSetting,
  in: "path",
  target: "user_id",
  type: "snowflake",
};

export const roleTargetField: BlockField = {
  ...roleTargetSetting,
  in: "path",
  target: "role_id",
  type: "snowflake",
};

// Path parameters of requests whose schema and editor input are generated.

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

export const flowChannelField: BlockField = {
  name: "channel_target",
  in: "path",
  target: "channel_id",
  type: "snowflake",
  label: "Channel",
  description:
    "ID of the channel. Leave empty to use the channel the flow runs in.",
  fallback: "channel",
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
      input: `condition_${kind}_base_value`,
      schema: templated(z.string(), baseValueDescription),
    },
    {
      name: "condition_allow_multiple",
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
      input: `condition_item_${kind}_mode`,
      schema: modeSchema,
    },
    {
      name: "condition_item_value",
      input: `condition_item_${kind}_value`,
      schema: templated(z.string(), valueDescription).optional(),
    },
  ];
}

// Message blocks send either an inline message or a saved template, see
// requireMessage.
export const messageDataField: BlockField = {
  name: "message_data",
  schema: messageDataSchema.optional(),
};

export const messageTemplateField: BlockField = {
  name: "message_template_id",
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
