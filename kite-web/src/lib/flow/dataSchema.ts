import { Edge, Node, NodeProps as XYNodeProps } from "@xyflow/react";
import z from "zod";
import { FlowNodeData } from "../types/flow.gen";

const numericRegex = /^[0-9]+$/;
const decimalRegex = /^[0-9]+(\.[0-9]+)?$/;
// A single placeholder, like {{arg('user').id}}.
const placeholderRegex = /^\{\{[^{}]+\}\}$/;

export interface FlowData {
  nodes: Node<NodeData>[];
  edges: Edge[];
}

export type NodeData = FlowNodeData & Record<string, unknown>;

export type NodeProps = XYNodeProps<Node<NodeData>>;

export type NodeType = Node<NodeData>;

// Fields that are evaluated as templates when the flow runs, so they can mix
// plain text with placeholders like {{user.mention}}. Tracked by the zod def
// because zod 3 has no other place for custom metadata.
const templatedDefs = new WeakSet<z.ZodTypeDef>();

export function templated<T extends z.ZodTypeAny>(
  schema: T,
  description: string
): T {
  const described = schema.describe(description);
  templatedDefs.add(described._def);
  return described;
}

export function isTemplated(def: z.ZodTypeDef) {
  return templatedDefs.has(def);
}

// A number or Discord ID, or a single placeholder that resolves to one.
function numericOrPlaceholder(description: string, regex = numericRegex) {
  const message = "Must be a number or ID, or a single {{ }} placeholder";
  return z
    .string()
    .regex(regex, message)
    .or(z.string().regex(placeholderRegex, message))
    .describe(description);
}

// Not enforced here, as the service doesn't validate flows of message
// components, so saved ones can hold other names.
const temporaryNameSchema = z
  .string()
  .max(32)
  .optional()
  .describe(
    "Stores the result of this block in a temporary variable that later blocks can read with {{var('name')}}. Lowercase letters, numbers and underscores only."
  );

export const auditLogReasonSchema = templated(
  z.string().max(512),
  "Reason shown in the server's audit log."
).optional();

const comparisonModeSchema = z.enum([
  "equal",
  "not_equal",
  "greater_than",
  "less_than",
  "greater_than_or_equal",
  "less_than_or_equal",
  "contains",
  "starts_with",
  "ends_with",
]);

export const nodeBaseDataSchema = z.object({
  custom_label: z
    .string()
    .optional()
    .describe(
      "Label shown on the block in the editor instead of its default title. It has no effect on what the flow does."
    ),
});

// Blocks without settings, like the else branch of a condition.
export const nodeEmptyDataSchema = z.object({});

export const nodeEntryCommandDataSchema = nodeBaseDataSchema.extend({
  name: z
    .string()
    .max(32)
    .min(1)
    .regex(
      /^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$/,
      "Must be only lowercase alphanumeric characters and underscores, and have at most 3 words"
    )
    .describe(
      "Name of the slash command. Up to three space separated words create subcommands, e.g. 'ticket open'."
    ),
  description: z
    .string()
    .max(100)
    .min(1)
    .describe("Description of the command shown in Discord."),
});

export const nodeOptionCommandArgumentDataSchema = nodeBaseDataSchema.extend({
  name: z
    .string()
    .max(32)
    .min(1)
    .regex(/^[a-z0-9_]+$/)
    .describe(
      "Name of the argument. Its value can be read with {{arg('name')}}."
    ),
  description: z
    .string()
    .max(100)
    .min(1)
    .describe("Description of the argument shown in Discord."),
  command_argument_type: z
    .enum([
      "string",
      "integer",
      "boolean",
      "user",
      "channel",
      "role",
      "mentionable",
      "number",
      "attachment",
    ])
    .describe("Type of value the user has to enter."),
  command_argument_required: z
    .boolean()
    .optional()
    .describe("Whether the user has to fill in the argument."),
  command_argument_min_value: z
    .number()
    .optional()
    .describe("Smallest allowed value for integer and number arguments."),
  command_argument_max_value: z
    .number()
    .optional()
    .describe("Largest allowed value for integer and number arguments."),
  command_argument_max_length: z
    .number()
    .optional()
    .describe("Maximum length for string arguments."),
  command_argument_choices: z
    .array(
      z.object({
        name: z.string().max(100).describe("Name shown to the user."),
        value: z
          .string()
          .max(100)
          .describe("Value the flow receives when the choice is picked."),
      })
    )
    .optional()
    .describe(
      "Fixed choices the user picks from, for string and integer arguments. Choices with an empty name or value are ignored."
    ),
});

export const nodeOptionCommandPermissionsSchema = nodeBaseDataSchema.extend({
  command_permissions: z
    .string()
    .regex(numericRegex)
    .describe(
      "Discord permission bitfield. Only members with all of these permissions can see and use the command."
    ),
});

export const nodeOptionCommandContextsSchema = nodeBaseDataSchema.extend({
  command_disabled_contexts: z
    .array(z.enum(["guild", "bot_dm", "private_channel"]))
    .optional()
    .describe(
      "Places where the command can't be used: servers, DMs with the bot, or other DMs and group DMs."
    ),
  command_disabled_integrations: z
    .array(z.enum(["guild_install", "user_install"]))
    .optional()
    .describe(
      "Install types the command isn't available for: installed to a server, or installed to a user."
    ),
});

export const nodeOptionEventFilterSchema = nodeBaseDataSchema.extend({
  event_filter_target: z
    .enum(["message_content", "user_id", "guild_id", "channel_id"])
    .describe("Property of the event to filter on."),
  event_filter_mode: z
    .enum(["equal", "not_equal", "contains", "starts_with", "ends_with"])
    .describe("How the property is compared to the filter value."),
  event_filter_value: z
    .string()
    .max(1000)
    .min(1)
    .describe(
      "Value to compare against. This is fixed text, placeholders aren't supported."
    ),
});

export const nodeEntryEventDataSchema = nodeBaseDataSchema.extend({
  event_type: z
    .enum([
      "message_create",
      "message_update",
      "message_delete",
      "guild_member_add",
      "guild_member_remove",
      "cron",
    ])
    .describe(
      "Discord event that triggers the flow, or cron for a scheduled flow."
    ),
  event_schedule_cron: z
    .string()
    .max(100)
    .optional()
    .describe(
      "Cron expression in UTC for scheduled flows, e.g. */5 * * * * for every five minutes."
    ),
  description: z
    .string()
    .max(100)
    .min(1)
    .describe("Description of what the event listener does."),
});

export const nodeEntryComponentButtonDataSchema = nodeEmptyDataSchema;

export const nodeMessageDataSchema = z
  .object({
    content: templated(
      z.string().max(2000),
      "Text content of the message."
    ).optional(),
    // Embeds are validated by the message editor, not here, as flows saved
    // before this was checked can hold incomplete ones.
    embeds: z
      .array(z.record(z.unknown()))
      .max(10)
      .optional()
      .describe(
        "Embeds shown below the content, as Discord embed objects (title, description, url, color, fields, author, footer, image, thumbnail). Their text supports placeholders."
      ),
    allowed_mentions: z
      .object({
        parse: z
          .array(z.enum(["users", "roles", "everyone"]))
          .optional()
          .describe("Kinds of mentions that ping."),
      })
      .optional()
      .describe(
        "Which mentions ping. If unset, only mentioned users are pinged."
      ),
  })
  // Components, flags and attachments are set through the message editor.
  .passthrough()
  .describe("The message to send.");

// Message blocks send either an inline message or a saved template.
function withMessage<T extends z.ZodRawShape>(shape: T) {
  return nodeBaseDataSchema
    .extend({
      ...shape,
      message_data: nodeMessageDataSchema.optional(),
      message_template_id: z
        .string()
        .optional()
        .describe(
          "ID of a saved message template to send instead of message_data."
        ),
      temporary_name: temporaryNameSchema,
    })
    .refine(
      (data) => !!data.message_data || !!data.message_template_id,
      "Either message_data or message_template_id is required"
    )
    .describe("Set either message_data or message_template_id.");
}

const channelTargetSchema = numericOrPlaceholder("ID of the channel.");
const messageTargetSchema = numericOrPlaceholder("ID of the message.");
const userTargetSchema = numericOrPlaceholder("ID of the user.");
const roleTargetSchema = numericOrPlaceholder("ID of the role.");
const guildTargetSchema = numericOrPlaceholder(
  "ID of the server. Defaults to the server the flow runs in."
);

const responseTargetMessage =
  "Must be an ID, '@original', or a single {{ }} placeholder";
const responseTargetSchema = z
  .string()
  .regex(numericRegex, responseTargetMessage)
  .or(z.string().regex(placeholderRegex, responseTargetMessage))
  .or(z.literal("@original"))
  .describe(
    "ID of the response message, or '@original' for the first response to the interaction."
  );

export const nodeActionResponseCreateDataSchema = withMessage({
  message_ephemeral: z
    .boolean()
    .optional()
    .describe(
      "Whether only the user who triggered the flow can see the response."
    ),
});

export const nodeActionResponseEditDataSchema = withMessage({
  message_target: responseTargetSchema,
});

export const nodeActionResponseDeleteDataSchema = nodeBaseDataSchema.extend({
  message_target: responseTargetSchema,
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionResponseDeferDataSchema = nodeBaseDataSchema.extend({
  message_ephemeral: z
    .boolean()
    .optional()
    .describe(
      "Whether only the user who triggered the flow can see the response that follows."
    ),
});

export const nodeSuspendResponseModalDataSchema = nodeBaseDataSchema.extend({
  modal_data: z
    .object({
      title: templated(z.string().max(45).min(1), "Title of the modal."),
      components: z
        .array(
          z.object({
            components: z
              .array(
                z.object({
                  custom_id: z
                    .string()
                    .max(100)
                    .min(1)
                    .describe(
                      "Identifier of the input. The submitted value can be read with {{input('custom_id')}}. This is fixed text, placeholders aren't supported."
                    ),
                  label: templated(
                    z.string().max(45).min(1),
                    "Label shown above the input."
                  ),
                  style: z
                    .literal(1)
                    .or(z.literal(2))
                    .describe("1 for a single line input, 2 for a paragraph."),
                  required: z
                    .boolean()
                    .optional()
                    .describe("Whether the input has to be filled in."),
                  min_length: z
                    .number()
                    .optional()
                    .describe("Minimum length of the entered text."),
                  max_length: z
                    .number()
                    .optional()
                    .describe("Maximum length of the entered text."),
                  value: templated(
                    z.string().max(4000).min(1),
                    "Value the input is pre-filled with."
                  ).optional(),
                  placeholder: templated(
                    z.string().max(4000).min(1),
                    "Text shown while the input is empty."
                  ).optional(),
                })
              )
              .min(1)
              .max(1)
              .describe("The text input of the row."),
          })
        )
        .min(1)
        .max(5)
        .describe("Rows of the modal, each holding one text input."),
    })
    .describe("The modal to show."),
});

export const nodeActionMessageCreateDataSchema = withMessage({
  channel_target: numericOrPlaceholder(
    "ID of the channel to send the message to."
  ),
});

export const nodeActionPrivateMessageCreateDataSchema = withMessage({
  user_target: numericOrPlaceholder("ID of the user to send the message to."),
});

export const nodeActionMessageEditDataSchema = withMessage({
  channel_target: channelTargetSchema,
  message_target: messageTargetSchema,
});

export const nodeActionMessageDeleteDataSchema = nodeBaseDataSchema.extend({
  channel_target: channelTargetSchema,
  message_target: messageTargetSchema,
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMessagePinDataSchema = nodeActionMessageDeleteDataSchema;

export const emojiDataSchema = z.object({
  id: z.string().optional().describe("ID of a custom emoji."),
  name: z
    .string()
    .min(1)
    .describe("Name of a custom emoji, or the unicode of a standard emoji."),
});

export const nodeActionMessageReactionCreateDataSchema =
  nodeBaseDataSchema.extend({
    channel_target: channelTargetSchema,
    message_target: messageTargetSchema,
    emoji_data: emojiDataSchema.describe("The emoji to react with."),
  });

export const nodeActionMessageReactionDeleteDataSchema =
  nodeBaseDataSchema.extend({
    channel_target: channelTargetSchema,
    message_target: messageTargetSchema,
    emoji_data: emojiDataSchema.describe(
      "The emoji to remove the reaction of."
    ),
  });

export const nodeActionMemberBanDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  user_target: userTargetSchema,
  member_ban_delete_message_duration_seconds: numericOrPlaceholder(
    "Delete the member's messages from this many seconds before the ban."
  ).optional(),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberUnbanDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  user_target: userTargetSchema,
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberKickDataSchema = nodeActionMemberUnbanDataSchema;

export const nodeActionMemberTimeoutDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  user_target: userTargetSchema,
  member_timeout_duration_seconds: numericOrPlaceholder(
    "How many seconds the member is timed out for."
  ),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberEditDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  user_target: userTargetSchema,
  member_data: z
    .object({
      nick: templated(z.string(), "New nickname of the member."),
    })
    .describe("The changes to make to the member."),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberRoleAddDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  user_target: userTargetSchema,
  role_target: roleTargetSchema,
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberRoleRemoveDataSchema =
  nodeActionMemberRoleAddDataSchema;

export const nodeActionMemberGetDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  user_target: userTargetSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionUserGetDataSchema = nodeBaseDataSchema.extend({
  user_target: userTargetSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionChannelGetDataSchema = nodeBaseDataSchema.extend({
  channel_target: channelTargetSchema,
  temporary_name: temporaryNameSchema,
});

export const channelDataSchema = z
  .object({
    name: templated(z.string().max(100).min(1), "Name of the channel."),
    type: z
      .number()
      .optional()
      .describe(
        "Discord channel type: 0 text, 2 voice, 4 category, 5 announcement, 13 stage, 15 forum, 16 media. For threads: 10 announcement, 11 public, 12 private."
      ),
    topic: templated(
      z.string().max(1000).min(1),
      "Topic of the channel."
    ).optional(),
    nsfw: z
      .boolean()
      .optional()
      .describe("Whether the channel is age-restricted."),
    bitrate: numericOrPlaceholder("Bitrate of a voice channel.").optional(),
    user_limit: numericOrPlaceholder(
      "Maximum number of users in a voice channel."
    ).optional(),
    position: numericOrPlaceholder(
      "Position of the channel in the channel list."
    ).optional(),
    parent: numericOrPlaceholder(
      "ID of the category the channel is in, or of the channel a thread is created in."
    ).optional(),
    permission_overwrites: z
      .array(
        z.object({
          id: numericOrPlaceholder("ID of the role or user."),
          type: z
            .literal(0)
            .or(z.literal(1))
            .optional()
            .describe("0 if id is a role, 1 if it's a user."),
          allow: numericOrPlaceholder("Permission bitfield to allow."),
          deny: numericOrPlaceholder("Permission bitfield to deny."),
        })
      )
      .optional()
      .describe("Permission overrides for roles and users."),
    invitable: z
      .boolean()
      .optional()
      .describe("Whether non-moderators can add members to a private thread."),
  })
  .describe("Settings of the channel, thread or forum post.");

export const nodeActionChannelCreateDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  channel_data: channelDataSchema,
  audit_log_reason: auditLogReasonSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionChannelEditDataSchema = nodeBaseDataSchema.extend({
  channel_target: channelTargetSchema,
  channel_data: channelDataSchema,
  audit_log_reason: auditLogReasonSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionChannelDeleteDataSchema = nodeBaseDataSchema.extend({
  channel_target: channelTargetSchema,
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionThreadCreateDataSchema = nodeBaseDataSchema.extend({
  message_target: numericOrPlaceholder(
    "ID of the message to start the thread from. Leave unset for a thread without a starter message."
  ).optional(),
  channel_data: channelDataSchema,
  audit_log_reason: auditLogReasonSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionThreadMemberAddDataSchema = nodeBaseDataSchema.extend({
  channel_target: numericOrPlaceholder("ID of the thread."),
  user_target: templated(z.string(), "ID of the user."),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionThreadMemberRemoveDataSchema =
  nodeActionThreadMemberAddDataSchema;

export const nodeActionForumPostCreateDataSchema = nodeBaseDataSchema.extend({
  channel_target: numericOrPlaceholder("ID of the forum channel."),
  channel_data: channelDataSchema,
  audit_log_reason: auditLogReasonSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionRoleGetDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
  role_target: roleTargetSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionGuildGetDataSchema = nodeBaseDataSchema.extend({
  guild_target: numericOrPlaceholder("ID of the server."),
  temporary_name: temporaryNameSchema,
});

export const nodeActionMessageGetDataSchema = nodeBaseDataSchema.extend({
  channel_target: numericOrPlaceholder(
    "ID of the channel the message is in. Defaults to the channel the flow runs in."
  ).optional(),
  message_target: messageTargetSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionRobloxUserGetDataSchema = nodeBaseDataSchema.extend({
  roblox_user_target: templated(
    z.string(),
    "ID or username of the Roblox user, depending on roblox_lookup_mode."
  ),
  roblox_lookup_mode: z
    .enum(["id", "username"])
    .describe("Whether roblox_user_target is an ID or a username."),
  temporary_name: temporaryNameSchema,
});

const variableIdSchema = z
  .string()
  .describe("ID of an existing stored variable.");

const variableScopeSchema = templated(
  z.string(),
  "Scope of the value, e.g. a user ID to store one value per user. Only for scoped variables."
).optional();

export const nodeActionVariableSetSchema = nodeBaseDataSchema.extend({
  variable_id: variableIdSchema,
  variable_scope: variableScopeSchema,
  variable_value: templated(z.string(), "Value to store."),
  variable_operation: z
    .enum(["overwrite", "append", "prepend", "increment", "decrement"])
    .describe(
      "How the value is combined with the stored one. increment and decrement add or subtract a number."
    ),
  temporary_name: temporaryNameSchema,
});

export const nodeActionVariableDeleteSchema = nodeBaseDataSchema.extend({
  variable_id: variableIdSchema,
  variable_scope: variableScopeSchema,
});

export const nodeActionVariableGetSchema = nodeBaseDataSchema.extend({
  variable_id: variableIdSchema,
  variable_scope: variableScopeSchema,
  temporary_name: temporaryNameSchema,
});

export const nodeActionVoiceChannelJoinDataSchema = nodeBaseDataSchema.extend({
  channel_target: numericOrPlaceholder("ID of the voice channel to join."),
  voice_self_mute: z
    .boolean()
    .optional()
    .describe("Whether the bot joins muted."),
  voice_self_deaf: z
    .boolean()
    .optional()
    .describe("Whether the bot joins deafened."),
});

export const nodeActionVoiceChannelLeaveDataSchema = nodeBaseDataSchema.extend({
  guild_target: guildTargetSchema.optional(),
});

export const nodeActionStatusSetDataSchema = nodeBaseDataSchema.extend({
  status_data: z
    .object({
      status: z
        .enum(["online", "idle", "dnd", "invisible"])
        .optional()
        .describe("Online status of the bot. Defaults to online."),
      activity_type: z
        .number()
        .optional()
        .describe(
          "Activity type: 0 Playing, 1 Streaming, 2 Listening, 3 Watching, 4 Custom, 5 Competing."
        ),
      activity_name: templated(
        z.string().min(1).max(128),
        "Text of the activity shown on the bot's profile."
      ),
      activity_url: templated(
        z.string(),
        "Stream URL, only used by the Streaming activity type."
      ).optional(),
    })
    // Validates the fields even before any was set, so their errors show up
    .default({ activity_name: "" })
    .describe("The status to set."),
});

export const nodeActionHttpRequestDataSchema = nodeBaseDataSchema.extend({
  http_request_data: z
    .object({
      url: templated(z.string().url(), "URL to send the request to."),
      method: z
        .enum(["GET", "POST", "PUT", "PATCH", "DELETE"])
        .describe("HTTP method of the request."),
      headers: z
        .array(
          z.object({
            key: z.string().describe("Name of the header."),
            value: templated(z.string(), "Value of the header."),
          })
        )
        .optional()
        .describe("Headers to send with the request."),
      body_json: z
        .record(z.unknown())
        .optional()
        .describe(
          "JSON body of the request. Placeholders in its string values are evaluated."
        ),
    })
    .describe("The request to send."),
  temporary_name: temporaryNameSchema,
});

const aiModelSchema = z
  .enum([
    "gpt-4.1",
    "gpt-4.1-mini",
    "gpt-4.1-nano",
    "gpt-5-nano",
    "gpt-4o-mini",
  ])
  .optional()
  .describe(
    "Model to use. Larger models cost more credits. Defaults to gpt-4o-mini."
  );

const aiMaxCompletionTokensSchema = numericOrPlaceholder(
  "Maximum number of tokens in the answer."
).optional();

export const nodeActionAiChatCompletionDataSchema = nodeBaseDataSchema.extend({
  ai_chat_completion_data: z
    .object({
      model: aiModelSchema,
      system_prompt: templated(
        z.string().max(2000),
        "Instructions for how the AI should behave."
      ).optional(),
      prompt: templated(
        z.string().max(2000).min(1),
        "Message the AI responds to."
      ),
      max_completion_tokens: aiMaxCompletionTokensSchema,
    })
    .describe("The prompt and model settings."),
  temporary_name: temporaryNameSchema,
});

export const nodeActionAiWebSearchCompletionDataSchema =
  nodeBaseDataSchema.extend({
    ai_chat_completion_data: z
      .object({
        model: aiModelSchema,
        prompt: templated(
          z.string().max(2000).min(1),
          "What to search the web for."
        ),
        max_completion_tokens: aiMaxCompletionTokensSchema,
      })
      .describe("The search query and model settings."),
    temporary_name: temporaryNameSchema,
  });

export const nodeActionExpressionEvaluateDataSchema = nodeBaseDataSchema.extend(
  {
    expression: templated(
      z
        .string()
        .max(2000)
        .refine((val) => !val.startsWith("{{"), {
          message:
            "In most cases, you don't need to use the double curly brackets around the expression here. Only use them if you want to include a placeholder in the expression.",
        }),
      "Expr language expression to evaluate, written without surrounding {{ }}, e.g. arg('a') + arg('b')."
    ),
    temporary_name: temporaryNameSchema,
  }
);

export const nodeActionRandomGenerateDataSchema = nodeBaseDataSchema.extend({
  random_min: numericOrPlaceholder("Smallest number that can be generated."),
  random_max: numericOrPlaceholder(
    "Upper bound of the generated number. The number is always below it."
  ),
  temporary_name: temporaryNameSchema,
});

export const nodeActionLogDataSchema = nodeBaseDataSchema.extend({
  log_level: z
    .enum(["debug", "info", "warn", "error"])
    .describe("Severity of the log entry."),
  log_message: templated(
    z.string().max(2000).min(1),
    "Text to write to the app's logs."
  ),
});

function conditionSchema(baseValueDescription: string) {
  return nodeBaseDataSchema.extend({
    condition_base_value: templated(z.string(), baseValueDescription),
    condition_allow_multiple: z
      .boolean()
      .optional()
      .describe(
        "Whether every matching branch runs. If unset, only the first matching branch runs."
      ),
  });
}

export const nodeConditionCompareDataSchema = conditionSchema(
  "Value that each branch compares against."
);

export const nodeConditionItemCompareDataSchema = nodeBaseDataSchema.extend({
  condition_item_mode: comparisonModeSchema.describe(
    "How the condition's base value is compared to this branch's value."
  ),
  condition_item_value: templated(
    z.string(),
    "Value to compare the base value with."
  ).optional(),
});

export const nodeConditionUserDataSchema = conditionSchema(
  "ID of the user that each branch checks."
);

export const nodeConditionItemUserDataSchema = nodeBaseDataSchema.extend({
  condition_item_mode: z
    .enum([
      "equal",
      "not_equal",
      "has_role",
      "not_has_role",
      "has_permission",
      "not_has_permission",
    ])
    .describe("What to check about the user."),
  condition_item_value: templated(
    z.string(),
    "User ID for equal and not_equal, role ID for has_role and not_has_role, or permission bitfield for has_permission and not_has_permission."
  ).optional(),
});

export const nodeConditionChannelDataSchema = conditionSchema(
  "ID of the channel that each branch checks."
);

export const nodeConditionRoleDataSchema = conditionSchema(
  "ID of the role that each branch checks."
);

// Channel and role conditions can only check for equality.
export const nodeConditionItemIdDataSchema = nodeBaseDataSchema.extend({
  condition_item_mode: z
    .enum(["equal", "not_equal"])
    .describe("Whether the base value has to match this branch's value."),
  condition_item_value: templated(
    z.string(),
    "ID to compare the base value with."
  ).optional(),
});

export const nodeControlErrorHandlerDataSchema = nodeBaseDataSchema.extend({
  temporary_name: temporaryNameSchema,
});

export const nodeControlLoopDataSchema = nodeBaseDataSchema.extend({
  loop_count: numericOrPlaceholder("How many times the loop runs."),
});

export const nodeControlSleepDataSchema = nodeBaseDataSchema.extend({
  sleep_duration_seconds: numericOrPlaceholder(
    "How many seconds to wait before continuing.",
    decimalRegex
  ),
});
