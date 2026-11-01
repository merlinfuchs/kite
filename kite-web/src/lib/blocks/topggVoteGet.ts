import { z } from "zod";
import { userTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

// Custom, as Top.gg answers 404 for a user without an active vote, which a
// request block would fail on.
export const topggVoteGet: BlockDefinition = {
  type: "action_topgg_vote_get",
  id_prefix: "vote",
  title: "Check Top.gg vote",
  description: "Check whether a user voted for your bot or server on Top.gg",
  icon: "vote",
  category: "Top.gg",
  credits: 1,
  strict_settings: true,
  requires: ["topgg"],
  fields: [userTargetSetting],
  result: {
    schema: z
      .object({
        voted: z
          .boolean()
          .describe(
            "Whether the user has an active vote, false if they never voted or can vote again"
          ),
        created_at: z
          .string()
          .describe(
            "When the user last voted, as an ISO 8601 timestamp, empty without an active vote"
          ),
        expires_at: z
          .string()
          .describe(
            "When the user can vote again, as an ISO 8601 timestamp, empty without an active vote"
          ),
        weight: z
          .number()
          .describe(
            "How many votes the last vote counted for, 2 on weekends, 0 without an active vote"
          ),
      })
      .describe("The user's vote"),
  },
  run: { kind: "custom" },
};
