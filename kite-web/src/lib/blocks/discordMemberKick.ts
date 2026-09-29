import { nodeActionMemberKickDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberKick: BlockDefinition = {
  type: "action_member_kick",
  title: "Kick member",
  description: "Kick a member from the server",
  icon: "user-round-minus",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMemberKickDataSchema,
  inputs: ["guild_target", "user_target", "audit_log_reason", "custom_label"],
  run: { kind: "custom" },
};
