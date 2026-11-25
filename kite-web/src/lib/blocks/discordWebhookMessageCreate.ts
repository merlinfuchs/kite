import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { onlyPlainSecrets } from "../flow/templateSecrets";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
import { BlockDefinition } from "./types";

export const discordWebhookMessageCreate: BlockDefinition = {
  type: "action_webhook_message_create",
  id_prefix: "hookmsg",
  title: "Send webhook message",
  description: "Sends a message through a webhook",
  icon: "send",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "webhook_url",
      secrets: true,
      schema: templated(
        z
          .string()
          .min(1)
          .refine(
            onlyPlainSecrets,
            "Secrets can only be used as {{secrets.NAME}}"
          ),
        "URL of the webhook, like https://discord.com/api/webhooks/ID/TOKEN. It can be built from the result of a Create webhook block, or be a secret like {{secrets.WEBHOOK_URL}}."
      ),
    },
    {
      name: "webhook_thread_id",
      schema: numericOrPlaceholder(
        "ID of a thread of the webhook's channel to send the message to. Leave unset to send it to the channel."
      ).optional(),
    },
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  result: { thing: "discord_message" },
  run: { kind: "custom" },
};
