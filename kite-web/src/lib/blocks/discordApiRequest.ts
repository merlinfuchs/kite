import { discordApiRequestDataSchema } from "../flow/dataSchema";
import { nodeActionDiscordApiRequestResultSchema } from "../flow/resultSchema";
import { requestBlocks } from ".";
import { BlockDefinition } from "./types";

export const discordApiRequest: BlockDefinition = {
  type: "action_discord_api_request",
  title: "Discord API Request",
  description:
    "Send a request to any endpoint of the Discord API, for things no other block does",
  icon: "braces",
  category: "API Requests",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "discord_api_request_data",
      type: "string",
      schema: () =>
        discordApiRequestDataSchema(
          requestBlocks()
            .filter((b) => b.run.integration === "discord" && !b.run.partial)
            .map((b) => `${b.run.operation}: ${b.type}`)
            .join(", ")
        ),
    },
  ],
  audit_log_reason: true,
  result: { schema: nodeActionDiscordApiRequestResultSchema },
  run: { kind: "custom" },
};
