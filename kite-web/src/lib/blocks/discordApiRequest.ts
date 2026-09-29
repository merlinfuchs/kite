import { nodeActionDiscordApiRequestDataSchema } from "../flow/dataSchema";
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
  schema: () =>
    nodeActionDiscordApiRequestDataSchema(
      requestBlocks()
        .filter((b) => b.run.integration === "discord")
        .map((b) => `${b.run.operation}: ${b.type}`)
        .join(", ")
    ),
  inputs: [
    "discord_api_request_data",
    "audit_log_reason",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionDiscordApiRequestResultSchema },
  run: { kind: "custom" },
};
