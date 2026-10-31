import { z } from "zod";
import { BlockDefinition } from "./types";

export const topggProjectGet: BlockDefinition = {
  type: "action_topgg_project_get",
  id_prefix: "topgg",
  title: "Get Top.gg listing",
  description: "Get the votes and reviews of your bot or server on Top.gg",
  icon: "trophy",
  category: "Top.gg",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "topgg",
    operation: "get_project",
    method: "GET",
    path: "/projects/@me",
  },
  fields: [],
  result: {
    schema: z
      .object({
        id: z.string().describe("Top.gg ID of the listing"),
        name: z.string().describe("Name of the bot or server"),
        headline: z.string().describe("Short description on the listing"),
        type: z.string().describe("What is listed: bot or server"),
        platform: z.string().describe("Platform of the listing, like discord"),
        tags: z.array(z.string()).describe("Tags of the listing"),
        votes: z.number().describe("Votes this month"),
        votes_total: z.number().describe("Votes of all time"),
        review_score: z.number().describe("Average review score"),
        review_count: z.number().describe("Number of reviews"),
      })
      .describe("The listing"),
  },
};
