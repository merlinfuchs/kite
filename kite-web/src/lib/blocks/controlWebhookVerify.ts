import { AnyZodObject, z } from "zod";
import { templated } from "../flow/dataSchema";
import { onlyPlainSecrets } from "../flow/templateSecrets";
import { errorColor } from "./colors";
import { BlockDefinition } from "./types";

// The services the block knows how they sign, and what their key is. Their
// verification is in the service's pkg/webhook/verify.go.
export const webhookVerifyPresets = [
  { value: "erlc", label: "ER:LC" },
  {
    value: "github",
    label: "GitHub",
    key: "The webhook secret, like {{secrets.GITHUB_SECRET}}.",
  },
  {
    value: "stripe",
    label: "Stripe",
    key: "The signing secret starting with whsec_, like {{secrets.STRIPE_SECRET}}.",
  },
  {
    value: "slack",
    label: "Slack",
    key: "The app's signing secret, like {{secrets.SLACK_SECRET}}.",
  },
  // Discord signs with the app's public key, which Kite knows.
  { value: "discord", label: "Discord" },
  {
    value: "standard_webhooks",
    label: "Standard Webhooks (Svix, Clerk, Resend)",
    key: "The signing secret starting with whsec_, like {{secrets.WEBHOOK_SECRET}}, or the public key starting with whpk_.",
  },
  {
    value: "custom",
    label: "Custom",
    key: "The secret for HMAC, like {{secrets.WEBHOOK_SECRET}}, or the public key for Ed25519 as hex, base64 or PEM.",
  },
] as const satisfies readonly { value: string; label: string; key?: string }[];

export const webhookVerifyAlgorithms = [
  { value: "hmac_sha256", label: "HMAC SHA-256" },
  { value: "hmac_sha1", label: "HMAC SHA-1" },
  { value: "hmac_sha512", label: "HMAC SHA-512" },
  { value: "ed25519", label: "Ed25519" },
] as const;

export const webhookVerifyEncodings = [
  { value: "hex", label: "Hex" },
  { value: "base64", label: "Base64" },
] as const;

function values<T extends string>(options: readonly { value: T }[]) {
  return options.map((o) => o.value) as [T, ...T[]];
}

export const controlWebhookVerify: BlockDefinition = {
  type: "control_webhook_verify",
  title: "Verify webhook signature",
  description:
    "Check that the webhook request was signed by the service that sent it",
  icon: "shield-check",
  category: "Conditions",
  contexts: ["event_webhook"],
  outputs: ["valid", "invalid"],
  output_labels: {
    valid: { label: "Valid" },
    invalid: { label: "Invalid", color: errorColor },
  },
  component: "branches",
  strict_settings: true,
  fields: [
    {
      name: "webhook_verify_preset",
      input: "webhook_verify",
      schema: z
        .enum(values(webhookVerifyPresets))
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
        "The signing secret, usually {{secrets.NAME}}, or the public key for Ed25519. ER:LC and Discord need none."
      ),
    },
    {
      name: "webhook_verify_algorithm",
      input: "webhook_verify",
      schema: z
        .enum(values(webhookVerifyAlgorithms))
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
        .enum(values(webhookVerifyEncodings))
        .optional()
        .describe("Encoding of a custom signature. Defaults to hex."),
    },
  ],
  refine: (schema: AnyZodObject) =>
    schema.superRefine((data, ctx) => {
      const preset = webhookVerifyPresets.find(
        (p) => p.value === data.webhook_verify_preset
      );
      // Presets without a key, like ER:LC and Discord, ignore one left from
      // another.
      const key = data.webhook_verify_key ?? "";
      if (preset && "key" in preset && !key) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Required",
          path: ["webhook_verify_key"],
        });
      } else if (preset && "key" in preset && !onlyPlainSecrets(key)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Secrets can only be used as {{secrets.NAME}}",
          path: ["webhook_verify_key"],
        });
      }
      if (preset?.value === "custom") {
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
