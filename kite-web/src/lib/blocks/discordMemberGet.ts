import { nodeActionMemberGetResultSchema } from "../flow/resultSchema";
import { guildTargetSetting, userTargetSetting } from "./fields";
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
  fields: [guildTargetSetting, userTargetSetting],
  result: { schema: nodeActionMemberGetResultSchema },
  run: { kind: "custom" },
};
