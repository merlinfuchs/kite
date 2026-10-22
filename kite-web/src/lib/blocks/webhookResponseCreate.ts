import { z } from "zod";
import {
  numericOrPlaceholder,
  placeholderRegex,
  templated,
} from "../flow/dataSchema";
import { BlockDefinition } from "./types";

// Like the service, which doesn't allow redirects or other statuses with a
// meaning beyond the response.
function isAllowedStatus(status: number) {
  return (status >= 200 && status <= 299) || (status >= 400 && status <= 599);
}

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
      ).refine(
        (v) => placeholderRegex.test(v) || isAllowedStatus(Number(v)),
        "Must be between 200 and 299 or between 400 and 599"
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
