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

export const flowValueSchema = z
  .object({
    type: z.literal("string"),
    value: z.string(),
  })
  .or(
    z.object({
      type: z.literal("number"),
      value: z.number(),
    })
  );

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

export const nodeActionResponseCreateDataSchema = nodeBaseDataSchema.extend({
  message_data: z.object({
    content: z.string().max(2000).min(1),
  }),
  message_ephemeral: z.boolean().optional(),
});

export const nodeActionResponseEditDataSchema = nodeBaseDataSchema.extend({
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex))
    .or(z.literal("@original")),
  message_data: z.object({
    content: z.string().max(2000).min(1),
  }),
});

export const nodeActionResponseDeleteDataSchema = nodeBaseDataSchema.extend({
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex))
    .or(z.literal("@original")),
});

export const nodeActionMessageCreateDataSchema = nodeBaseDataSchema.extend({
  message_data: z.object({
    content: z.string().max(2000).min(1),
  }),
});

export const nodeActionMessageEditDataSchema = nodeBaseDataSchema.extend({
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  message_data: z.object({
    content: z.string().max(2000).min(1),
  }),
});

export const nodeActionMessageDeleteDataSchema = nodeBaseDataSchema.extend({
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
  member_ban_delete_message_duration: z.string().regex(numericRegex),
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
  member_timeout_duration: z.string().regex(numericRegex),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionChannelCreateDataSchema = nodeBaseDataSchema.extend({
  channel_data: z.object({
    name: z.string(),
  }),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionChannelEditDataSchema = nodeBaseDataSchema.extend({
  channel_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  channel_data: z.object({
    name: z.string(),
  }),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionChannelDeleteDataSchema = nodeBaseDataSchema.extend({
  channel_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionThreadCreateDataSchema = nodeBaseDataSchema.extend({
  message_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex))
    .optional(),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionRoleCreateDataSchema = nodeBaseDataSchema.extend({
  role_data: z.object({
    name: z.string(),
  }),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionRoleEditDataSchema = nodeBaseDataSchema.extend({
  role_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  role_data: z.object({
    name: z.string(),
  }),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionRoleDeleteDataSchema = nodeBaseDataSchema.extend({
  role_target: z
    .string()
    .regex(numericRegex)
    .or(z.string().regex(variableRegex)),
  audit_log_reason: auditLogReasonSchema,
});

export const nodeActionVariableSetDataSchema = nodeBaseDataSchema.extend({
  variable_name: z.string(),
  variable_value: flowValueSchema,
});

export const nodeActionVariableDeleteDataSchema = nodeBaseDataSchema.extend({
  variable_name: z.string(),
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
  condition_base_value: flowValueSchema,
  condition_allow_multiple: z.boolean().optional(),
});

export const nodeConditionItemCompareDataSchema = nodeBaseDataSchema.extend({
  condition_item_value: flowValueSchema,
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
