import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordWebhookCreate: BlockDefinition = {
  type: "action_webhook_create",
  id_prefix: "hook",
  title: "Create webhook",
  description: "Create a webhook in a channel",
  icon: "webhook",
  category: "Channels",
  permissions: ["MANAGE_WEBHOOKS"],
  audit_log_reason: true,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "create_webhook",
  },
  fields: [
    {
      name: "channel_target",
      target: "channel_id",
      label: "Channel",
      description:
        "ID of the channel to create the webhook in. Leave empty to use the channel the flow runs in. Threads and forums can't have webhooks of their own.",
      fallback: "channel",
    },
    {
      name: "name",
      label: "Name",
      description:
        "Name of the webhook, 1 to 80 characters. It can't contain 'discord' or 'clyde'.",
      required: true,
    },
  ],
  result: {
    schema: z
      .object({
        id: z.string().describe("The ID of the webhook"),
        token: z
          .string()
          .describe(
            "The secret token of the webhook. Anyone who has it can send messages through the webhook, so don't post it in a channel."
          ),
        name: z.string().describe("The name of the webhook"),
        channel_id: z
          .string()
          .describe("The ID of the channel the webhook was created in"),
      })
      .describe("The created webhook"),
  },
};
