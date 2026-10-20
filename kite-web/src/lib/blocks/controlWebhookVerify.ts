import { AnyZodObject, z } from "zod";
import { templated } from "../flow/dataSchema";
import { onlyPlainSecrets } from "../flow/templateSecrets";
import { BlockDefinition } from "./types";

export const webhookVerifyPresets = [
  "custom",
  "erlc",
  "github",
  "stripe",
  "slack",
  "discord",
  "standard_webhooks",
] as const;

export const controlWebhookVerify: BlockDefinition = {
  type: "control_webhook_verify",
  title: "Verify webhook signature",
  description:
    "Check that the webhook request was signed by the service that sent it",
  icon: "shield-check",
  category: "Conditions",
  contexts: ["event_webhook"],
  outputs: ["valid", "invalid"],
  component: "control_webhook_verify",
  strict_settings: true,
  fields: [
    {
      name: "webhook_verify_preset",
      input: "webhook_verify",
      schema: z
        .enum(webhookVerifyPresets)
        .describe(
          "Service that signs the requests, which decides the rest, or custom to set the algorithm, message and signature."
        ),
    },
    {
      name: "webhook_verify_key",
      input: "webhook_verify",
      secrets: true,
      schema: templated(
        z.string().optional(),
        "The signing secret, usually {{secrets.NAME}}, or the public key for Ed25519. ER:LC needs none."
      ),
    },
    {
      name: "webhook_verify_algorithm",
      input: "webhook_verify",
      schema: z
        .enum(["hmac_sha256", "hmac_sha1", "hmac_sha512", "ed25519"])
        .optional()
        .describe("Signing algorithm of a custom signature."),
    },
    {
      name: "webhook_verify_message",
      input: "webhook_verify",
      schema: templated(
        z.string().optional(),
        "What a custom signature signs, e.g. {{webhook.body}}. Spaces count."
      ),
    },
    {
      name: "webhook_verify_signature",
      input: "webhook_verify",
      schema: templated(
        z.string().optional(),
        "The custom signature, e.g. {{webhook.headers['x-signature']}}."
      ),
    },
    {
      name: "webhook_verify_signature_encoding",
      input: "webhook_verify",
      schema: z
        .enum(["hex", "base64"])
        .optional()
        .describe("Encoding of a custom signature. Defaults to hex."),
    },
  ],
  refine: (schema: AnyZodObject) =>
    schema.superRefine((data, ctx) => {
      const preset = data.webhook_verify_preset;
      const key = data.webhook_verify_key ?? "";
      if (preset !== "erlc" && !key) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Required",
          path: ["webhook_verify_key"],
        });
      }
      if (!onlyPlainSecrets(key)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Secrets can only be used as {{secrets.NAME}}",
          path: ["webhook_verify_key"],
        });
      }
      if (preset === "custom") {
        for (const field of [
          "webhook_verify_message",
          "webhook_verify_signature",
        ]) {
          if (!data[field]) {
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: "Required",
              path: [field],
            });
          }
        }
      }
    }),
  run: { kind: "custom" },
};
