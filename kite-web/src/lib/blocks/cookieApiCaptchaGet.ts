import { z } from "zod";
import { BlockDefinition } from "./types";

export const cookieApiCaptchaGet: BlockDefinition = {
  type: "action_cookie_api_captcha_get",
  title: "Check captcha",
  description: "Check if a captcha from Cookie API was solved",
  icon: "shield-question",
  category: "Cookie API",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "get_captcha",
    method: "GET",
    path: "/api/security/captcha/get-captcha",
  },
  fields: [
    {
      name: "captcha_id",
      in: "query",
      type: "string",
      label: "Captcha ID",
      description: "ID of the captcha, from the Create captcha block.",
      required: true,
    },
  ],
  result: {
    schema: z
      .object({
        solved: z.string().describe('"YES" if the user solved it, "NO" if not'),
        solved_at: z
          .string()
          .describe("Unix timestamp of when it was solved, 0 if it wasn't"),
      })
      .describe("The captcha"),
  },
};
