import { nodeActionMemberRoleRemoveDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberRoleRemove: BlockDefinition = {
  type: "action_member_role_remove",
  title: "Remove role from member",
  description: "Remove a role from a member",
  icon: "bookmark-minus",
  category: "Roles",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMemberRoleRemoveDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "role_target",
    "audit_log_reason",
    "custom_label",
  ],
  run: { kind: "custom" },
};
