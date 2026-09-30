import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { nodeActionRobloxUserGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const robloxUserGet: BlockDefinition = {
  type: "action_roblox_user_get",
  title: "Get Roblox User",
  description: "Get a Roblox user by ID or username",
  icon: "gamepad",
  category: "Roblox",
  requires: ["roblox"],
  credits: 1,
  allow_unknown_settings: true,
  fields: [
    {
      name: "roblox_user_target",
      type: "string",
      schema: templated(
        z.string(),
        "ID or username of the Roblox user, depending on roblox_lookup_mode."
      ),
    },
    {
      name: "roblox_lookup_mode",
      type: "string",
      schema: z
        .enum(["id", "username"])
        .describe("Whether roblox_user_target is an ID or a username."),
    },
  ],
  result: { schema: nodeActionRobloxUserGetResultSchema },
  run: { kind: "custom" },
};
