import { AnyZodObject, ZodTypeAny, z } from "zod";
import {
  auditLogReasonSchema,
  decimalRegex,
  NodeData,
  nodeBaseDataSchema,
  discordApiParamFormats,
  placeholderRegex,
  templated,
  temporaryNameSchema,
} from "../flow/dataSchema";
import {
  BlockDefinition,
  BlockField,
  BlockRequestChoice,
  BlockFieldType,
  RequestBlockDefinition,
  RequestChoiceBlockDefinition,
} from "./types";
import { aiChatCompletion } from "./aiChatCompletion";
import { aiWebSearch } from "./aiWebSearch";
import { controlConditionChannel } from "./controlConditionChannel";
import { controlConditionCompare } from "./controlConditionCompare";
import { controlConditionItemChannel } from "./controlConditionItemChannel";
import { controlConditionItemCompare } from "./controlConditionItemCompare";
import { controlConditionItemElse } from "./controlConditionItemElse";
import { controlConditionItemRole } from "./controlConditionItemRole";
import { controlConditionItemUser } from "./controlConditionItemUser";
import { controlConditionRole } from "./controlConditionRole";
import { controlConditionUser } from "./controlConditionUser";
import { controlErrorHandler } from "./controlErrorHandler";
import { controlLoop } from "./controlLoop";
import { controlLoopEach } from "./controlLoopEach";
import { controlLoopEnd } from "./controlLoopEnd";
import { controlLoopExit } from "./controlLoopExit";
import { controlSleep } from "./controlSleep";
import { controlWebhookVerify } from "./controlWebhookVerify";
import { discordApiRequest } from "./discordApiRequest";
import { discordChannelCreate } from "./discordChannelCreate";
import { discordChannelDelete } from "./discordChannelDelete";
import { discordChannelEdit } from "./discordChannelEdit";
import { discordChannelGet } from "./discordChannelGet";
import { discordEntryCommand } from "./discordEntryCommand";
import { discordEntryComponentButton } from "./discordEntryComponentButton";
import { discordEntryEvent } from "./discordEntryEvent";
import { discordForumPostCreate } from "./discordForumPostCreate";
import { discordGuildGet } from "./discordGuildGet";
import { discordInviteCreate } from "./discordInviteCreate";
import { discordMemberBan } from "./discordMemberBan";
import { discordMemberEdit } from "./discordMemberEdit";
import { discordMemberGet } from "./discordMemberGet";
import { discordMemberKick } from "./discordMemberKick";
import { discordMemberRoleAdd } from "./discordMemberRoleAdd";
import { discordMemberRoleRemove } from "./discordMemberRoleRemove";
import { discordMemberTimeout } from "./discordMemberTimeout";
import { discordMemberUnban } from "./discordMemberUnban";
import { discordMessageBulkDelete } from "./discordMessageBulkDelete";
import { discordMessageCreate } from "./discordMessageCreate";
import { discordMessageDelete } from "./discordMessageDelete";
import { discordMessageEdit } from "./discordMessageEdit";
import { discordMessageGet } from "./discordMessageGet";
import { discordMessageList } from "./discordMessageList";
import { discordMessagePin } from "./discordMessagePin";
import { discordMessageReactionCreate } from "./discordMessageReactionCreate";
import { discordMessageReactionDelete } from "./discordMessageReactionDelete";
import { discordMessageUnpin } from "./discordMessageUnpin";
import { discordOptionCommandArgument } from "./discordOptionCommandArgument";
import { discordOptionCommandContexts } from "./discordOptionCommandContexts";
import { discordOptionCommandCooldown } from "./discordOptionCommandCooldown";
import { discordOptionCommandPermissions } from "./discordOptionCommandPermissions";
import { discordOptionEventFilter } from "./discordOptionEventFilter";
import { discordPollCreate } from "./discordPollCreate";
import { discordPrivateMessageCreate } from "./discordPrivateMessageCreate";
import { discordResponseCreate } from "./discordResponseCreate";
import { discordResponseDefer } from "./discordResponseDefer";
import { discordResponseDelete } from "./discordResponseDelete";
import { discordResponseEdit } from "./discordResponseEdit";
import { discordRoleCreate } from "./discordRoleCreate";
import { discordRoleGet } from "./discordRoleGet";
import { discordStatusSet } from "./discordStatusSet";
import { discordSuspendResponseModal } from "./discordSuspendResponseModal";
import { discordThreadCreate } from "./discordThreadCreate";
import { discordThreadMemberAdd } from "./discordThreadMemberAdd";
import { discordThreadMemberRemove } from "./discordThreadMemberRemove";
import { discordTypingTrigger } from "./discordTypingTrigger";
import { discordUserGet } from "./discordUserGet";
import { discordVoiceChannelJoin } from "./discordVoiceChannelJoin";
import { discordVoiceChannelLeave } from "./discordVoiceChannelLeave";
import { erlcCommandRun } from "./erlcCommandRun";
import { erlcServerGet } from "./erlcServerGet";
import { expressionEvaluate } from "./expressionEvaluate";
import { httpRequest } from "./httpRequest";
import { log } from "./log";
import { popcatFunGet } from "./popcatFunGet";
import { popcatImageCreate } from "./popcatImageCreate";
import { popcatSearch } from "./popcatSearch";
import { popcatTextTransform } from "./popcatTextTransform";
import { randomGenerate } from "./randomGenerate";
import { robloxUserGet } from "./robloxUserGet";
import { variableDelete } from "./variableDelete";
import { variableGet } from "./variableGet";
import { variableSet } from "./variableSet";
import { webhookResponseCreate } from "./webhookResponseCreate";

// Every block, in the order of the block explorer, then the blocks that aren't
// in it, like entries and the branches of conditions. The flow AI's catalog
// follows this order. File
// names start with the integration a block mainly acts on, if any.
export const blockDefinitions: BlockDefinition[] = [
  discordOptionCommandArgument,
  discordOptionCommandPermissions,
  discordOptionCommandContexts,
  discordOptionCommandCooldown,
  discordOptionEventFilter,
  discordResponseCreate,
  discordResponseEdit,
  discordResponseDelete,
  discordResponseDefer,
  discordSuspendResponseModal,
  webhookResponseCreate,
  discordMessageCreate,
  discordMessageEdit,
  discordMessageDelete,
  discordMessageGet,
  discordPrivateMessageCreate,
  discordTypingTrigger,
  discordMessageReactionCreate,
  discordMessageReactionDelete,
  discordMessagePin,
  discordMessageUnpin,
  discordPollCreate,
  discordMessageList,
  discordMessageBulkDelete,
  discordMemberBan,
  discordMemberUnban,
  discordMemberKick,
  discordMemberTimeout,
  discordMemberEdit,
  discordMemberGet,
  discordUserGet,
  discordMemberRoleAdd,
  discordMemberRoleRemove,
  discordRoleGet,
  discordRoleCreate,
  discordGuildGet,
  discordChannelCreate,
  discordChannelEdit,
  discordChannelDelete,
  discordChannelGet,
  discordThreadCreate,
  discordForumPostCreate,
  discordThreadMemberAdd,
  discordThreadMemberRemove,
  discordInviteCreate,
  discordVoiceChannelJoin,
  discordVoiceChannelLeave,
  discordStatusSet,
  variableSet,
  variableDelete,
  variableGet,
  robloxUserGet,
  erlcServerGet,
  erlcCommandRun,
  popcatFunGet,
  popcatTextTransform,
  popcatSearch,
  popcatImageCreate,
  aiChatCompletion,
  aiWebSearch,
  httpRequest,
  discordApiRequest,
  expressionEvaluate,
  randomGenerate,
  log,
  controlConditionCompare,
  controlConditionUser,
  controlConditionChannel,
  controlConditionRole,
  controlWebhookVerify,
  controlLoop,
  controlLoopExit,
  controlErrorHandler,
  controlSleep,
  discordEntryCommand,
  discordEntryEvent,
  discordEntryComponentButton,
  controlConditionItemCompare,
  controlConditionItemUser,
  controlConditionItemChannel,
  controlConditionItemRole,
  controlConditionItemElse,
  controlLoopEach,
  controlLoopEnd,
];

const definitionsByType = new Map(blockDefinitions.map((b) => [b.type, b]));

export function getBlockDefinition(type: string | undefined) {
  return type ? definitionsByType.get(type) : undefined;
}

// The integrations a block needs: the one its request goes to and the ones
// it requires.
export function blockIntegrations(block: BlockDefinition) {
  const ids = [...(block.requires ?? [])];
  if (
    (block.run.kind === "request" || block.run.kind === "request_choice") &&
    !ids.includes(block.run.integration)
  ) {
    ids.unshift(block.run.integration);
  }
  return ids;
}

// The blocks that send a request to an integration.
export function requestBlocks() {
  return blockDefinitions.filter(isRequestBlock);
}

// A test checks that every block with a request run has fields and credits.
export function isRequestBlock(
  block: BlockDefinition
): block is RequestBlockDefinition {
  return block.run.kind === "request";
}

// The blocks that send one of several requests to an integration.
export function requestChoiceBlocks() {
  return blockDefinitions.filter(isRequestChoiceBlock);
}

export function isRequestChoiceBlock(
  block: BlockDefinition
): block is RequestChoiceBlockDefinition {
  return block.run.kind === "request_choice";
}

// The operation picked in a request choice block, if any.
export function getBlockOperation(block: BlockDefinition, data: NodeData) {
  if (block.run.kind !== "request_choice") return;
  const id = data[block.run.setting];
  return block.run.operations.find((o) => o.id === id);
}

const formats: Record<BlockFieldType, [RegExp, string] | null> = {
  snowflake: discordApiParamFormats.snowflake,
  snowflake_list: [
    /^[0-9]+([,\s]+[0-9]+)*$/,
    "Must be IDs separated by commas",
  ],
  integer: discordApiParamFormats.integer,
  boolean: discordApiParamFormats.boolean,
  string: null,
  emoji: null,
  seconds: [decimalRegex, "Must be a number of seconds"],
  seconds_until: [decimalRegex, "Must be a number of seconds"],
};

const numberTypes: BlockFieldType[] = ["integer", "seconds", "seconds_until"];

function fieldSchema(field: BlockField) {
  // Emojis are objects, which only a block's own schema describes.
  if (!field.type || field.type === "emoji") {
    throw new Error(`Field ${field.name} needs its own schema`);
  }
  const type = field.type;

  let text = z.string();
  if (field.max_length) text = text.max(field.max_length);

  const withChecks = text.superRefine((value, ctx) => {
    if (!value) {
      if (field.required) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Required" });
      }
      return;
    }

    const [format, message] = formats[type] ?? [];
    // Placeholders are only checked when the flow runs.
    if (value.includes("{{")) {
      if (format && !placeholderRegex.test(value) && field.type !== "string") {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: `${message}, or a single {{ }} placeholder`,
        });
      }
      return;
    }
    if (format && !format.test(value)) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, message });
      return;
    }

    if (numberTypes.includes(type)) {
      const n = Number(value);
      if (field.min !== undefined && n < field.min) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: `Must be at least ${field.min}`,
        });
      }
      if (field.max !== undefined && n > field.max) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: `Must be at most ${field.max}`,
        });
      }
    }
    if (field.type === "snowflake_list") {
      const count = value.split(/[,\s]+/).filter(Boolean).length;
      if (field.min !== undefined && count < field.min) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: `Must be at least ${field.min} IDs`,
        });
      }
      if (field.max !== undefined && count > field.max) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: `Must be at most ${field.max} IDs`,
        });
      }
    }
  });

  // Lists of IDs can also be stored as a list, whose items can be
  // placeholders.
  const listOrText =
    field.type === "snowflake_list"
      ? withChecks.or(
          z
            .array(
              z
                .string()
                .regex(
                  /^([0-9]+|\{\{[^{}]+\}\})$/,
                  "Must be IDs or single {{ }} placeholders"
                )
            )
            .min(field.min ?? 0)
            .max(field.max ?? Infinity)
        )
      : withChecks;

  // Values are stored as text, so the flow AI needs to know the format.
  const described = templated(
    listOrText,
    field.type === "boolean"
      ? `${field.description} Either "true" or "false".`
      : field.description!
  );
  // Numbers and booleans work too, as long as they have the right format.
  const schema = z.preprocess(
    (v) => (typeof v === "number" || typeof v === "boolean" ? String(v) : v),
    described
  );
  return field.required ? schema : schema.optional();
}

// The schema of a block, generated from its fields.
export function blockDataSchema(block: BlockDefinition) {
  const fields = block.fields ?? [];
  const names = fields.map((f) => f.name).join(", ");

  let base: AnyZodObject =
    block.custom_label === false ? z.object({}) : nodeBaseDataSchema;
  if (block.strict_settings) {
    base = base.strict(`Unknown setting. This block's settings are: ${names}`);
  }

  const schema = base.extend({
    ...(block.run.kind === "request_choice" && {
      [block.run.setting]: operationSettingSchema(block.run),
    }),
    ...Object.fromEntries(fields.map((f) => [f.name, settingSchema(f)])),
    ...(block.audit_log_reason && { audit_log_reason: auditLogReasonSchema }),
    ...(block.result && { temporary_name: temporaryNameSchema }),
  });
  const refined = block.refine ? block.refine(schema) : schema;
  return block.run.kind === "request_choice"
    ? requireOperationFields(refined, block.run)
    : refined;
}

// The setting that picks the operation of a request choice block. The flow
// AI reads the operations and the fields they send from its description.
function operationSettingSchema(run: BlockRequestChoice) {
  const ids = run.operations.map((o) => o.id) as [string, ...string[]];
  const list = run.operations
    .map((o) => {
      const sent = o.fields
        .map((f) => `${f.name}${f.required ? "" : " (optional)"}`)
        .join(", ");
      return `"${o.id}": ${o.description}${sent ? ` Sends ${sent}.` : ""}`;
    })
    .join("\n");
  return z
    .enum(ids, { errorMap: () => ({ message: "Invalid input" }) })
    .describe(
      `What the block does, which decides the settings it sends:\n${list}`
    );
}

// Fields are optional in the schema, as what's required depends on the
// operation.
function requireOperationFields(schema: ZodTypeAny, run: BlockRequestChoice) {
  return schema.superRefine((data: NodeData, ctx) => {
    const op = run.operations.find((o) => o.id === data[run.setting]);
    for (const field of op?.fields ?? []) {
      const value = data[field.name];
      if (field.required && (value === undefined || value === "")) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: [field.name],
          message: "Required",
        });
      }
    }
  });
}

function settingSchema(field: BlockField) {
  if (typeof field.schema === "function") return field.schema();
  return field.schema ?? fieldSchema(field);
}

// The editor inputs of a block. Fields without a schema are edited by a
// generic input, named "field:" and the field's name.
export function blockDataFields(block: BlockDefinition) {
  // Request choices show the fields of the picked operation in one input.
  const inputs =
    block.run.kind === "request_choice"
      ? ["request_choice"]
      : (block.fields ?? []).map((f) =>
          f.schema ? f.input ?? f.name : `field:${f.name}`
        );
  return [
    ...new Set(inputs),
    ...(block.audit_log_reason ? ["audit_log_reason"] : []),
    ...(block.result ? ["temporary_name"] : []),
    ...(block.custom_label === false ? [] : ["custom_label"]),
  ];
}
