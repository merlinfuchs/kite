import z from "zod";

export interface NodeData {
  custom_label?: string;
  name?: string;
  description?: string;
  text?: string;
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
