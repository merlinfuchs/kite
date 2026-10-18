import { z } from "zod";
import { BlockDefinition } from "./types";

export const cookieApiMinecraftUserGet: BlockDefinition = {
  type: "action_cookie_api_minecraft_user_get",
  title: "Verify Minecraft player",
  description: "Get the Minecraft player a Cookie API verification code is for",
  icon: "pickaxe",
  category: "Cookie API",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "get_minecraft_user",
    method: "GET",
    path: "/api/minecraft/get-user",
  },
  fields: [
    {
      name: "minecraft_code",
      in: "query",
      target: "code",
      type: "string",
      label: "Code",
      description:
        "The code the player got when joining verify.cookie-api.com in Minecraft.",
      required: true,
    },
  ],
  result: {
    schema: z
      .object({
        player_id: z.string().describe("UUID of the player"),
        player_name: z.string().describe("Name of the player"),
        edition: z.string().describe('"JAVA" or "BEDROCK"'),
      })
      .describe("The verified player"),
  },
};
