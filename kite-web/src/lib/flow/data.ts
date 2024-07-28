import { Edge, Node, NodeProps as XYNodeProps } from "@xyflow/react";
import z from "zod";
import { FlowNodeData } from "../types/flow.gen";

const numericRegex = /^[0-9]$/;

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

export const nodeActionMessageCreateDataSchema = nodeBaseDataSchema.extend({
  message_data: z.object({
    content: z.string().max(2000).min(1),
  }),
  result_variable_name: z.string().optional(),
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
