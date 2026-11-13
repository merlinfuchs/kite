import { z } from "zod";
import { BlockDefinition } from "./types";

export const cookieApiCaptchaCreate: BlockDefinition = {
  type: "action_cookie_api_captcha_create",
  id_prefix: "captcha",
  title: "Create captcha",
  description: "Create a captcha page with Cookie API for a user to solve",
  icon: "shield-check",
  category: "Cookie API",
  strict_settings: true,
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "create_captcha",
  },
  fields: [
    {
      name: "captcha_provider",
      label: "Provider",
      description: "Who shows the captcha.",
      options: [
        { value: "Cloudflare", label: "Cloudflare Turnstile" },
        { value: "Google", label: "Google reCAPTCHA" },
      ],
    },
    {
      name: "captcha_color",
      target: "color",
      label: "Color",
      description: "Accent color of the page as a hex code, like #4EAF54.",
    },
  ],
  result: {
    schema: z
      .object({
        captcha_id: z
          .string()
          .describe("ID of the captcha, to check later if it was solved"),
        url: z.string().describe("Link to the page the user solves it on"),
      })
      .describe("The created captcha, which expires after an hour"),
  },
};
