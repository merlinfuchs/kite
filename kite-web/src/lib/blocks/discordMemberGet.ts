import { guildTargetSchema, userTargetSchema } from "../flow/dataSchema";
import { nodeActionMemberGetResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordMemberGet: BlockDefinition = {
  type: "action_member_get",
  title: "Get member",
  description: "Get a member by ID",
  icon: "user-round-search",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  allow_unknown_settings: true,
  fields: [
    {
      name: "guild_target",
      type: "snowflake",
      schema: guildTargetSchema.optional(),
    },
    { name: "user_target", type: "snowflake", schema: userTargetSchema },
  ],
  result: { schema: nodeActionMemberGetResultSchema },
  run: { kind: "custom" },
};
