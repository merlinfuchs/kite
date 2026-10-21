import { z } from "zod";
import { discordApiRequestDataSchema } from "../flow/dataSchema";
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
  result: {
    schema: z
      .unknown()
      .describe(
        "The JSON the endpoint returns, as in Discord's API docs, e.g. result('id').code for a created invite or result('id')[0].content for a list of messages. Empty for endpoints that return nothing."
      ),
  },
  run: { kind: "custom" },
};
