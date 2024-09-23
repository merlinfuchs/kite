import { Edge, Node, NodeProps as XYNodeProps } from "@xyflow/react";
import z from "zod";
import { FlowNodeData } from "../types/flow.gen";

const numericRegex = /^[0-9]+$/;
const variableRegex = /^\{\{[a-z0-9_.]+\}\}$/;

export interface FlowData {
  nodes: Node<NodeData>[];
  edges: Edge[];
}

export type NodeData = FlowNodeData & Record<string, unknown>;

export type NodeProps = XYNodeProps<Node<NodeData>>;

export type NodeType = Node<NodeData>;

export const auditLogReasonSchema = z.string().max(512).optional();

export const nodeBaseDataSchema = z.object({
  custom_label: z.string().optional(),
});

export const nodeEntryCommandDataSchema = nodeBaseDataSchema.extend({
  name: z
    .string()
    .max(32)
    .min(1)
    .regex(
      /^[a-zA-Z0-9_]+$/,
      "Must be only alphanumeric characters and underscores"
    ),
  description: z.string().max(100).min(1),
});

export const nodeOptionCommandArgumentDataSchema = nodeBaseDataSchema.extend({
  name: z.string().max(32).min(1),
  description: z.string().max(100).min(1),
  command_argument_type: z
    .literal("string")
    .or(z.literal("integer"))
    .or(z.literal("boolean"))
    .or(z.literal("user"))
    .or(z.literal("channel"))
    .or(z.literal("role"))
    .or(z.literal("mentionable"))
    .or(z.literal("number"))
    .or(z.literal("attachment")),
  command_argument_required: z.boolean().optional(),
});

export const nodeOptionCommandPermissionsSchema = nodeBaseDataSchema.extend({
  command_permissions: z.string().regex(numericRegex),
});

export const nodeOptionCommandContextsSchema = nodeBaseDataSchema.extend({
  command_disabled_contexts: z
    .array(
      z
        .literal("guild")
        .or(z.literal("bot_dm"))
        .or(z.literal("private_channel"))
    )
    .optional(),
});

export const nodeOptionEventFilterSchema = nodeBaseDataSchema.extend({
  event_filter_target: z.literal("message_content"),
  event_filter_expression: z.string().max(1000).min(1),
});

export const nodeEntryEventDataSchema = nodeBaseDataSchema.extend({
  event_type: z.string(),
});

export const nodeEntryComponentButtonDataSchema = nodeBaseDataSchema.extend({});

export const nodeActionResponseCreateDataSchema = nodeBaseDataSchema
  .extend({
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
    message_ephemeral: z.boolean().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionResponseEditDataSchema = nodeBaseDataSchema
  .extend({
    message_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(variableRegex))
      .or(z.literal("@original")),
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionResponseDeleteDataSchema = nodeBaseDataSchema.extend({
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex))
    .or(z.literal("@original")),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMessageCreateDataSchema = nodeBaseDataSchema
  .extend({
    channel_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(variableRegex)),
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionMessageEditDataSchema = nodeBaseDataSchema
  .extend({
    channel_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(variableRegex)),
    message_target: z
      .string()
      .regex(numericRegex)
      .or(z.string().regex(variableRegex)),
    message_data: z
      .object({
        content: z.string().max(2000).min(1),
      })
      .optional(),
    message_template_id: z.string().optional(),
  })
  .refine(
    (data) => !!data.message_data || !!data.message_template_id,
    "Either message_data or message_template_id is required"
  );

export const nodeActionMessageDeleteDataSchema = nodeBaseDataSchema.extend({
  channel_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberBanDataSchema = nodeBaseDataSchema.extend({
  member_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  member_ban_delete_message_duration_seconds: z.string().regex(numericRegex),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberUnbanDataSchema = nodeBaseDataSchema.extend({
  member_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberKickDataSchema = nodeBaseDataSchema.extend({
  member_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberTimeoutDataSchema = nodeBaseDataSchema.extend({
  member_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  member_timeout_duration_seconds: z.string().regex(numericRegex),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionMemberEditDataSchema = nodeBaseDataSchema.extend({
  member_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  member_data: z.object({
    nick: z.string(),
  }),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionHttpRequestDataSchema = nodeBaseDataSchema.extend({
  http_request_data: z.object({
    url: z.string().url(),
    method: z
      .literal("GET")
      .or(z.literal("POST"))
      .or(z.literal("PUT"))
      .or(z.literal("PATCH"))
      .or(z.literal("DELETE")),
  }),
});

export const nodeActionLogDataSchema = nodeBaseDataSchema.extend({
  log_level: z
    .literal("debug")
    .or(z.literal("info"))
    .or(z.literal("warn"))
    .or(z.literal("error")),
  log_message: z.string().max(2000).min(1),
});

export const nodeConditionCompareDataSchema = nodeBaseDataSchema.extend({
  condition_base_value: z.string(),
  condition_allow_multiple: z.boolean().optional(),
});

export const nodeConditionItemCompareDataSchema = nodeBaseDataSchema.extend({
  condition_item_value: z.string().optional(),
  condition_item_mode: z
    .literal("equal")
    .or(z.literal("not_equal"))
    .or(z.literal("greater_than"))
    .or(z.literal("less_than"))
    .or(z.literal("greater_than_or_equal"))
    .or(z.literal("less_than_or_equal"))
    .or(z.literal("contains")),
});

export const nodeControlLoopDataSchema = nodeBaseDataSchema.extend({
  loop_count: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
});

export const nodeControlSleepDataSchema = nodeBaseDataSchema.extend({
  sleep_duration_seconds: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
});
