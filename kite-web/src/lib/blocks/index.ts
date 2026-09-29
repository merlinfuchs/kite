import { z } from "zod";
import {
  auditLogReasonSchema,
  nodeBaseDataSchema,
  discordApiParamFormats,
  placeholderRegex,
  templated,
  temporaryNameSchema,
} from "../flow/dataSchema";
import { BlockDefinition, BlockField, RequestBlockDefinition } from "./types";
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
import { discordUserGet } from "./discordUserGet";
import { discordVoiceChannelJoin } from "./discordVoiceChannelJoin";
import { discordVoiceChannelLeave } from "./discordVoiceChannelLeave";
import { expressionEvaluate } from "./expressionEvaluate";
import { httpRequest } from "./httpRequest";
import { log } from "./log";
import { randomGenerate } from "./randomGenerate";
import { robloxUserGet } from "./robloxUserGet";
import { variableDelete } from "./variableDelete";
import { variableGet } from "./variableGet";
import { variableSet } from "./variableSet";

// Every block, in the order of the block explorer, then the blocks that aren't
// in it, like entries and the branches of conditions. The flow AI's catalog
// follows this order. File
// names start with the integration a block mainly acts on, if any.
export const blockDefinitions: BlockDefinition[] = [
  discordOptionCommandArgument,
  discordOptionCommandPermissions,
  discordOptionCommandContexts,
  discordOptionEventFilter,
  discordResponseCreate,
  discordResponseEdit,
  discordResponseDelete,
  discordResponseDefer,
  discordSuspendResponseModal,
  discordMessageCreate,
  discordMessageEdit,
  discordMessageDelete,
  discordMessageGet,
  discordPrivateMessageCreate,
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
  controlLoop,
  controlLoopExit,
  controlErrorHandler,
  controlSleep,
  discordEntryCommand,
  discordEntryEvent,
  discordEntryComponentButton,
  discordForumPostCreate,
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
  if (block.run.kind === "request" && !ids.includes(block.run.integration)) {
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

const formats: Record<BlockField["type"], [RegExp, string] | null> = {
  snowflake: discordApiParamFormats.snowflake,
  snowflake_list: [
    /^[0-9]+([,\s]+[0-9]+)*$/,
    "Must be IDs separated by commas",
  ],
  integer: discordApiParamFormats.integer,
  boolean: discordApiParamFormats.boolean,
  string: null,
  emoji: null,
  seconds_until: [/^[0-9]+$/, "Must be a number of seconds"],
};

function fieldSchema(field: BlockField) {
  let text = z.string();
  if (field.max_length) text = text.max(field.max_length);

  const withChecks = text.superRefine((value, ctx) => {
    if (!value) {
      if (field.required) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: "Required" });
      }
      return;
    }

    const [format, message] = formats[field.type] ?? [];
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

    if (field.type === "integer") {
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
      : field.description
  );
  // Numbers and booleans work too, as long as they have the right format.
  const schema = z.preprocess(
    (v) => (typeof v === "number" || typeof v === "boolean" ? String(v) : v),
    described
  );
  return field.required ? schema : schema.optional();
}

export function blockDataSchema(block: RequestBlockDefinition) {
  const names = block.fields.map((f) => f.name).join(", ");

  // Strict, as settings with a wrong name would otherwise be dropped silently.
  return nodeBaseDataSchema
    .strict(`Unknown setting. This block's settings are: ${names}`)
    .extend({
      ...Object.fromEntries(block.fields.map((f) => [f.name, fieldSchema(f)])),
      ...(block.audit_log_reason && { audit_log_reason: auditLogReasonSchema }),
      ...(block.result && { temporary_name: temporaryNameSchema }),
    });
}

export function blockDataFields(block: RequestBlockDefinition) {
  return [
    "block_fields",
    ...(block.audit_log_reason ? ["audit_log_reason"] : []),
    ...(block.result ? ["temporary_name"] : []),
    "custom_label",
  ];
}
