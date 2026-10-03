import { nodeActionBotStatsGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordBotStatsGet: BlockDefinition = {
  type: "action_bot_stats_get",
  title: "Get bot stats",
  description: "Get the server count, member count, uptime and latency",
  icon: "chart-column",
  category: "Bot",
  requires: ["discord"],
  strict_settings: true,
  credits: 1,
  fields: [],
  result: { schema: nodeActionBotStatsGetResultSchema },
  run: { kind: "custom" },
};
