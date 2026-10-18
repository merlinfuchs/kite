import { z } from "zod";
import { BlockDefinition } from "./types";

export const cookieApiCardCreate: BlockDefinition = {
  type: "action_cookie_api_card_create",
  title: "Generate card",
  description: "Generate an image like a welcome card with Cookie API",
  icon: "id-card",
  category: "Cookie API",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "create_card",
    method: "POST",
    path: "/api/cards/card-builder/build",
  },
  fields: [
    {
      name: "card_data",
      in: "body",
      type: "json_object",
      label: "Card",
      description:
        "The card as JSON, designed at https://trolensdesign.github.io/CardBuilder-cookie-api/. Placeholders in its texts are filled in.",
      required: true,
    },
  ],
  result: {
    schema: z
      .object({
        url: z.string().describe("URL of the card's image"),
      })
      .describe("The generated card"),
  },
};
