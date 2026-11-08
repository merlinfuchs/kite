import { z } from "zod";
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
import { onlyPlainSecrets, templateSecretNames } from "../flow/templateSecrets";
import { BlockField, SettingsSchema } from "./types";

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

// Where requests send the shared settings.
const guildPath = {
  in: "path",
  target: "guild_id",
  type: "snowflake",
  fallback: "guild",
} as const;

const channelPath = {
  in: "path",
  target: "channel_id",
  type: "snowflake",
} as const;

export const guildTargetField: BlockField = {
  ...guildTargetSetting,
  ...guildPath,
};

export const channelTargetField: BlockField = {
  ...channelTargetSetting,
  ...channelPath,
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

// The same path parameters for blocks whose schema and editor input are
// generated.

export const guildField: BlockField = {
  name: "guild_target",
  ...guildPath,
  label: "Server",
  description:
    "ID of the server. Leave empty to use the server the flow runs in.",
};

export const flowChannelField: BlockField = {
  name: "channel_target",
  ...channelPath,
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
  modeSchema: z.ZodType,
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

// A secret in a condition's value only works on its own, like
// {{secrets.TOKEN}}, and with equal and not equal. Otherwise whoever controls
// the base value could guess it. Ignoring case would let a token with the
// wrong case through.
export function secretsOnlyWithEquality(schema: SettingsSchema) {
  return schema.superRefine((data, ctx) => {
    const value = data.condition_item_value ?? "";
    if (templateSecretNames(value).length === 0) return;

    if (!onlyPlainSecrets(value)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Secrets can only be compared as {{secrets.NAME}}",
        path: ["condition_item_value"],
      });
    } else if (
      data.condition_item_mode !== "equal" &&
      data.condition_item_mode !== "not_equal"
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Secrets can only be compared with Equal or Not Equal",
        path: ["condition_item_value"],
      });
    }

    if (data.condition_item_ignore_case) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Secrets can't be compared with Ignore Case",
        path: ["condition_item_ignore_case"],
      });
    }
  });
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

export function requireMessage(schema: SettingsSchema) {
  return schema
    .refine(
      (data) => !!data.message_data || !!data.message_template_id,
      "Either message_data or message_template_id is required"
    )
    .describe("Set either message_data or message_template_id.");
}

// The settings of a role, for creating one or editing one, where empty
// settings keep what the role has.
export function roleDataFields(edit: boolean): BlockField[] {
  const describe = (description: string, created = "") =>
    `${description} ${
      edit ? "Leave empty to keep the current one." : created
    }`.trim();
  return [
    {
      name: "name",
      in: "body",
      type: "string",
      label: "Name",
      description: describe("Name of the role.", "Defaults to 'new role'."),
      max_length: 100,
    },
    {
      name: "permissions",
      in: "body",
      type: "string",
      label: "Permissions",
      description: describe(
        "Permissions of the role.",
        "Defaults to the ones of @everyone."
      ),
      widget: "permissions",
    },
    {
      name: "color",
      in: "body",
      type: "color",
      label: "Color",
      description: describe(
        `Color of the role, a hex code like #ff0000 or a number like 16711680. ${
          edit ? "0 removes it." : ""
        }`.trim(),
        "Defaults to no color."
      ),
      min: 0,
      max: 16777215,
    },
    {
      name: "hoist",
      in: "body",
      type: "boolean",
      label: "Show Separately",
      description: describe(
        "Whether members with the role are shown separately in the member list."
      ),
    },
    {
      name: "mentionable",
      in: "body",
      type: "boolean",
      label: "Mentionable",
      description: describe("Whether everyone can mention the role."),
    },
  ];
}
