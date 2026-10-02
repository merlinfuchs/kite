import { guildTargetSetting, userTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberTimeoutRemove: BlockDefinition = {
  type: "action_member_timeout_remove",
  title: "Remove member timeout",
  description: "Remove the timeout of a member in the server",
  icon: "message-circle",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  fields: [guildTargetSetting, userTargetSetting],
  run: { kind: "custom" },
};
