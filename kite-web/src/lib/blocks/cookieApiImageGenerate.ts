import { z } from "zod";
import { BlockDefinition } from "./types";

export const cookieApiImageGenerate: BlockDefinition = {
  type: "action_cookie_api_image_generate",
  title: "Generate AI image",
  description: "Generate an image from a description with Cookie API",
  icon: "image-plus",
  category: "Cookie API",
  credits: 1,
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "generate_image",
    method: "POST",
    path: "/api/ai/generate-image",
  },
  fields: [
    {
      name: "image_prompt",
      in: "body",
      target: "prompt",
      type: "string",
      label: "Description",
      description: "What the image should show.",
      required: true,
    },
  ],
  result: {
    schema: z
      .object({
        url: z.string().describe("URL of the image, which is kept for 7 days"),
      })
      .describe("The generated image"),
  },
};
