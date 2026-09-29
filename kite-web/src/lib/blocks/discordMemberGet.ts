import { nodeActionMemberGetDataSchema } from "../flow/dataSchema";
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
  schema: nodeActionMemberGetDataSchema,
  inputs: ["guild_target", "user_target", "temporary_name", "custom_label"],
  result: { schema: nodeActionMemberGetResultSchema },
  run: { kind: "custom" },
};
