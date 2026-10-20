import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const webhookResponseCreate: BlockDefinition = {
  type: "action_webhook_response_create",
  title: "Respond to webhook",
  description:
    "Answer the request that triggered the webhook with a status and body",
  icon: "reply",
  category: "Responses",
  contexts: ["event_webhook"],
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "webhook_response_status",
      schema: numericOrPlaceholder(
        "HTTP status code, between 200 and 299 or between 400 and 599."
      ),
    },
    {
      name: "webhook_response_content_type",
      schema: z
        .enum(["application/json", "text/plain"])
        .optional()
        .describe("Content type of the body. Defaults to text/plain."),
    },
    {
      name: "webhook_response_body",
      schema: templated(
        z
          .string()
          .max(64 * 1024)
          .optional(),
        "Body of the response, e.g. JSON when the content type is application/json."
      ),
    },
  ],
  run: { kind: "custom" },
};
