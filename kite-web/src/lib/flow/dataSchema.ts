import { Edge, Node, NodeProps as XYNodeProps } from "@xyflow/react";
import z from "zod";
import { ApiParam } from "../integrations/api";
import { FlowNodeData } from "../types/flow.gen";
import { aiModelTierValues, resolveAiModel } from "./aiModels";
import {
  getDiscordApiOperation,
  similarDiscordApiOperations,
} from "./discordApi";

export const numericRegex = /^[0-9]+$/;
export const decimalRegex = /^[0-9]+(\.[0-9]+)?$/;
// A single placeholder, like {{arg('user').id}}.
export const placeholderRegex = /^\{\{[^{}]+\}\}$/;

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

// Fields that refer to something only the user can create in the app, like a
// stored variable, so the AI leaves them for the user to pick.
const userPickedDefs = new WeakSet<z.ZodTypeDef>();

export function userPicked<T extends z.ZodTypeAny>(
  schema: T,
  description: string
): T {
  const described = schema.describe(description);
  userPickedDefs.add(described._def);
  return described;
}

export function isUserPicked(def: z.ZodTypeDef) {
  return userPickedDefs.has(def);
}

// Whether the setting at path of a block's settings is picked by the user.
// They are all top-level settings.
export function isUserPickedSetting(
  schema: z.ZodTypeAny,
  path: (string | number)[]
) {
  const object = unwrap(schema);
  if (path.length !== 1 || !(object instanceof z.ZodObject)) return false;
  for (
    let s: z.ZodTypeAny | undefined = object.shape[path[0]];
    s;
    s = inner(s)
  ) {
    if (isUserPicked(s._def)) return true;
  }
  return false;
}

function unwrap(schema: z.ZodTypeAny) {
  let s = schema;
  for (let next = inner(s); next; next = inner(s)) s = next;
  return s;
}

function inner(schema: z.ZodTypeAny): z.ZodTypeAny | undefined {
  if (
    schema instanceof z.ZodOptional ||
    schema instanceof z.ZodNullable ||
    schema instanceof z.ZodDefault
  ) {
    return schema._def.innerType;
  }
  if (schema instanceof z.ZodEffects) return schema._def.schema;
}

// A number or Discord ID, or a single placeholder that resolves to one.
export function numericOrPlaceholder(
  description: string,
  regex = numericRegex
) {
  const message = "Must be a number or ID, or a single {{ }} placeholder";
  return z
    .string()
    .regex(regex, message)
    .or(z.string().regex(placeholderRegex, message))
    .describe(description);
}

// Not enforced here, as the service doesn't validate flows of message
// components, so saved ones can hold other names.
export const temporaryNameSchema = z
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

// Channel and role conditions can only check for equality.
export const idConditionModeSchema = z
  .enum(["equal", "not_equal"])
  .describe("Whether the base value has to match this branch's value.");

export const comparisonModeSchema = z.enum([
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

export const messageDataSchema = z
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
    // Validated by the message editor like embeds.
    components: z
      .array(z.record(z.unknown()))
      .optional()
      .describe(
        "Buttons and select menus, as Discord action rows. Each one that isn't a link button adds the output component_<id> to the block."
      ),
  })
  // Flags and attachments are set through the message editor.
  .passthrough()
  .describe("The message to send.");

export const channelTargetSchema = numericOrPlaceholder("ID of the channel.");
export const messageTargetSchema = numericOrPlaceholder("ID of the message.");
export const userTargetSchema = numericOrPlaceholder("ID of the user.");
export const roleTargetSchema = numericOrPlaceholder("ID of the role.");
export const guildTargetSchema = numericOrPlaceholder(
  "ID of the server. Defaults to the server the flow runs in."
);

const responseTargetMessage =
  "Must be an ID, '@original', or a single {{ }} placeholder";
export const responseTargetSchema = z
  .string()
  .regex(numericRegex, responseTargetMessage)
  .or(z.string().regex(placeholderRegex, responseTargetMessage))
  .or(z.literal("@original"))
  .describe(
    "ID of the response message, or '@original' for the first response to the interaction."
  );

export const emojiDataSchema = z.object({
  id: z.string().optional().describe("ID of a custom emoji."),
  name: z
    .string()
    .min(1)
    .describe("Name of a custom emoji, or the unicode of a standard emoji."),
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

export const variableIdSchema = userPicked(
  z.string(),
  "ID of an existing stored variable."
);

export const variableScopeSchema = templated(
  z.string(),
  "Scope of the value, e.g. a user ID to store one value per user. Only for scoped variables."
).optional();

const discordApiParamSchema = z.object({
  key: z.string().describe("Name of the parameter."),
  value: templated(z.string(), "Value of the parameter."),
});

const operationDescription =
  "operationId of the endpoint in Discord's OpenAPI spec, e.g. create_message or list_messages. The spec's names can differ from Discord's docs, e.g. Modify Guild is update_guild.";

const discordApiRequestBaseSchema = z.object({
  operation: z.string().describe(operationDescription),
  path_params: z
    .array(discordApiParamSchema)
    .optional()
    .describe(
      "Values for the parameters in the endpoint's path, e.g. channel_id. IDs can also be a placeholder that resolves to a user, channel or other Discord object."
    ),
  query: z
    .array(discordApiParamSchema)
    .optional()
    .describe("Query parameters. Only the ones the endpoint has are allowed."),
  body_json: z
    .record(z.unknown())
    .or(z.array(z.unknown()))
    .optional()
    .describe(
      "JSON body of the request, an object or for some endpoints a list. Placeholders in its string values are evaluated. A string that is a single placeholder keeps the type of its result, e.g. a number or list."
    ),
});

// Formats of typed Discord API parameters, which can also be a placeholder.
export const discordApiParamFormats: Record<string, [RegExp, string]> = {
  snowflake: [numericRegex, "Must be a number or ID"],
  integer: [/^-?[0-9]+$/, "Must be a whole number"],
  number: [/^-?[0-9]+(\.[0-9]+)?$/, "Must be a number"],
  boolean: [/^(true|false)$/, "Must be true or false"],
};

// dedicated lists the endpoints that have their own block, like
// "list_messages: action_message_list", which the flow AI should use instead.
export function discordApiRequestDataSchema(dedicated: string) {
  return discordApiRequestBaseSchema
    .extend({
      operation: z
        .string()
        .describe(
          `${operationDescription} These endpoints have their own block, use it instead: ${dedicated}.`
        ),
    })
    .superRefine(refineDiscordApiRequest)
    .describe("The Discord API request to send. The bot's token is added.");
}

function refineDiscordApiRequest(
  data: z.infer<typeof discordApiRequestBaseSchema>,
  ctx: z.RefinementCtx
) {
  const op = getDiscordApiOperation(data.operation);
  if (!op) {
    const similar = similarDiscordApiOperations(data.operation);
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["operation"],
      message: `Unknown endpoint. Similar ones: ${similar.join(", ")}`,
    });
    return;
  }

  const checkParams = (
    field: "path_params" | "query",
    declared: ApiParam[]
  ) => {
    const values = new Map(data[field]?.map((p) => [p.key, p.value]));
    for (const p of declared) {
      const value = values.get(p.name);
      if (value === undefined ? p.required : !value) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: [field, p.name],
          message: `${p.name} is required`,
        });
        continue;
      }

      const [format, message] = discordApiParamFormats[p.type] ?? [];
      if (
        value &&
        format &&
        !format.test(value) &&
        !placeholderRegex.test(value)
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: [field, p.name],
          message: `${message}, or a single {{ }} placeholder`,
        });
      }
    }
    // The service ignores leftover path parameters, but not query parameters.
    if (field === "path_params") return;
    for (const key of Array.from(values.keys())) {
      if (!declared.some((p) => p.name === key)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: [field, key],
          message: `The endpoint has no parameter ${key}`,
        });
      }
    }
  };
  checkParams("path_params", op.path_params);
  checkParams("query", op.query_params);

  if (data.body_json && !op.has_body) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["body_json"],
      message: "The endpoint doesn't take a body",
    });
  }
}

export const aiModelSchema = z
  .preprocess(resolveAiModel, z.enum(aiModelTierValues).optional())
  .describe(
    "Model tier to use. Larger tiers are more capable and cost more credits. Defaults to small."
  );

export const aiMaxCompletionTokensSchema = numericOrPlaceholder(
  "Maximum number of tokens in the answer."
).optional();
