import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordStatusSet: BlockDefinition = {
  type: "action_status_set",
  title: "Set status",
  description: "Change the status and activity of the bot",
  icon: "activity",
  category: "Bot",
  requires: ["discord"],
  premium_feature: "rotating_status",
  credits: 1,
  fields: [
    {
      name: "status_data",
      type: "string",
      schema: z
        .object({
          status: z
            .enum(["online", "idle", "dnd", "invisible"])
            .optional()
            .describe("Online status of the bot. Defaults to online."),
          activity_type: z
            .number()
            .optional()
            .describe(
              "Activity type: 0 Playing, 1 Streaming, 2 Listening, 3 Watching, 4 Custom, 5 Competing."
            ),
          activity_name: templated(
            z.string().min(1).max(128),
            "Text of the activity shown on the bot's profile."
          ),
          activity_url: templated(
            z.string(),
            "Stream URL, only used by the Streaming activity type."
          ).optional(),
        })
        // Validates the fields even before any was set, so their errors show up
        .default({ activity_name: "" })
        .describe("The status to set."),
    },
  ],
  run: { kind: "custom" },
};
