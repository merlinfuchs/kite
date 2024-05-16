import { Edge, Node } from "reactflow";
import z from "zod";

export interface FlowData {
  nodes: Node<NodeData>[];
  edges: Edge[];
}

export interface NodeData {
  custom_label?: string;
  name?: string;
  description?: string;
  text?: string;
  log_level?: string;
  log_message?: string;

  condition_type?: string;
  condition_base_value?: string;
  condition_multiple?: boolean;
  condition_item_type?: string;
  condition_item_mode?: string;
  condition_item_value?: string;
}

export const nodeBaseDataSchema = z.object({
  custom_label: z.string().optional(),
});

export const nodeOptionDataSchema = nodeBaseDataSchema.extend({
  name: z.string().max(100).min(3),
  description: z.string().max(100).min(3),
});

export const nodeCommandDataSchema = nodeBaseDataSchema.extend({
  name: z.string().max(100).min(3),
  description: z.string().max(100).min(3),
});

export const nodeActionDataSchema = nodeBaseDataSchema.extend({
  text: z.string().max(2000).min(1),
});

export const nodeActionLogDataSchema = nodeBaseDataSchema.extend({
  log_level: z
    .literal("debug")
    .or(z.literal("info"))
    .or(z.literal("warn"))
    .or(z.literal("error")),
  log_message: z.string().max(2000).min(1),
});

export const nodeConditionDataSchema = nodeBaseDataSchema.extend({
  condition_type: z.literal("comparison"),
  condition_base_value: z.string(),
  condition_multiple: z.boolean(),
});

export const nodeConditionItemCompareDataSchema = nodeBaseDataSchema.extend({
  condition_item_value: z.string(),
  condition_item_mode: z.literal("equal").or(z.literal("not_equal")),
});
