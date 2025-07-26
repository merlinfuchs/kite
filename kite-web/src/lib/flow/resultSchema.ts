import { z } from "zod";

export const messageResultSchema = z.object({
  content: z.string().describe("The content of the message"),
});

export const nodeActionMessageCreateResultSchema = messageResultSchema;
