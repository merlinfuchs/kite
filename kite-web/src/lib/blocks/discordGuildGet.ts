import { numericOrPlaceholder } from "../flow/dataSchema";
import { nodeActionGuildGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordGuildGet: BlockDefinition = {
  type: "action_guild_get",
  title: "Get server",
  description: "Get a server / guild by ID",
  icon: "server",
  category: "Servers",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "guild_target",
      type: "snowflake",
      schema: numericOrPlaceholder("ID of the server."),
    },
  ],
  result: { schema: nodeActionGuildGetResultSchema },
  run: { kind: "custom" },
};
