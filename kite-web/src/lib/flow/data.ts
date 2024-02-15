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
}

export const nodeOptionDataSchema = z.object({
  custom_label: z.string().optional(),
  name: z.string().max(100).min(3),
  description: z.string().max(100).min(3),
});

export const nodeCommandDataSchema = z.object({
  custom_label: z.string().optional(),
  name: z.string().max(100).min(3),
  description: z.string().max(100).min(3),
});

export const nodeActionDataSchema = z.object({
  custom_label: z.string().optional(),
  text: z.string().max(2000).min(1),
});

export const nodeActionLogDataSchema = z.object({
  custom_label: z.string().optional(),
  log_level: z
    .literal("debug")
    .or(z.literal("info"))
    .or(z.literal("warn"))
    .or(z.literal("error")),
  log_message: z.string().max(2000).min(1),
});
