import { z } from "zod";
import {
  auditLogReasonSchema,
  nodeBaseDataSchema,
  discordApiParamFormats,
  placeholderRegex,
  templated,
  temporaryNameSchema,
} from "../flow/dataSchema";
import { discordBulkDeleteMessages } from "./discordBulkDeleteMessages";
import { discordCreateInvite } from "./discordCreateInvite";
import { discordCreateRole } from "./discordCreateRole";
import { discordListMessages } from "./discordListMessages";
import { BlockDefinition, BlockField } from "./types";

// File names start with the integration a block mainly acts on, if any.
export const blockDefinitions: BlockDefinition[] = [
  discordListMessages,
  discordBulkDeleteMessages,
  discordCreateInvite,
  discordCreateRole,
];

const definitionsByType = new Map(blockDefinitions.map((b) => [b.type, b]));

export function getBlockDefinition(type: string | undefined) {
  return type ? definitionsByType.get(type) : undefined;
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

export function blockDataSchema(block: BlockDefinition) {
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

export function blockDataFields(block: BlockDefinition) {
  return [
    "block_fields",
    ...(block.audit_log_reason ? ["audit_log_reason"] : []),
    ...(block.result ? ["temporary_name"] : []),
    "custom_label",
  ];
}
