import { BlockDefinition } from "./types";

export const topggMetricsUpdate: BlockDefinition = {
  type: "action_topgg_metrics_update",
  id_prefix: "topgg",
  title: "Post Top.gg server count",
  description: "Show how many servers your bot is in on its Top.gg listing",
  icon: "trending-up",
  category: "Top.gg",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "topgg",
    operation: "update_metrics",
    method: "PATCH",
    path: "/projects/@me/metrics",
  },
  fields: [
    {
      name: "topgg_server_count",
      in: "body",
      target: "server_count",
      type: "integer",
      min: 0,
      label: "Server count",
      description:
        "Number of servers the bot is in, usually {{app.guild_count}}. Post it from a scheduled event listener, e.g. every hour.",
      required: true,
    },
    {
      name: "topgg_shard_count",
      in: "body",
      target: "shard_count",
      type: "integer",
      min: 1,
      label: "Shard count",
      description: "Number of shards of the bot. Leave empty if it has none.",
    },
  ],
};
