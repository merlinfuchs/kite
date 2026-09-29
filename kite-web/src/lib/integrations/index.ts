import { z } from "zod";
import {
  auditLogReasonSchema,
  nodeBaseDataSchema,
  numericRegex,
  placeholderRegex,
  templated,
  temporaryNameSchema,
} from "../flow/dataSchema";
import { discord } from "./discord";
import { Integration, IntegrationBlock, IntegrationBlockField } from "./types";

export const integrations: Integration[] = [discord];

export const integrationBlocks = integrations.flatMap((i) => i.blocks);

const blocksByType = new Map(integrationBlocks.map((b) => [b.type, b]));

export function getIntegrationBlock(type: string | undefined) {
  return type ? blocksByType.get(type) : undefined;
}

const formats: Record<IntegrationBlockField["type"], [RegExp, string] | null> =
  {
    snowflake: [numericRegex, "Must be an ID"],
    snowflake_list: [
      /^[0-9]+([,\s]+[0-9]+)*$/,
      "Must be IDs separated by commas",
    ],
    integer: [/^-?[0-9]+$/, "Must be a whole number"],
    boolean: [/^(true|false)$/, "Must be true or false"],
    string: null,
  };

function fieldSchema(field: IntegrationBlockField) {
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

  // Values are stored as text, so the flow AI needs to know the format.
  const described = templated(
    withChecks,
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

export function integrationBlockDataSchema(block: IntegrationBlock) {
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

export function integrationBlockDataFields(block: IntegrationBlock) {
  return [
    "integration_fields",
    ...(block.audit_log_reason ? ["audit_log_reason"] : []),
    ...(block.result ? ["temporary_name"] : []),
    "custom_label",
  ];
}
