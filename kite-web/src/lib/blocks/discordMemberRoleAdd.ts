import { nodeActionMemberRoleAddDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMemberRoleAdd: BlockDefinition = {
  type: "action_member_role_add",
  title: "Add role to member",
  description: "Add a role to a member",
  icon: "bookmark-plus",
  category: "Roles",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMemberRoleAddDataSchema,
  inputs: [
    "guild_target",
    "user_target",
    "role_target",
    "audit_log_reason",
    "custom_label",
  ],
  run: { kind: "custom" },
};
