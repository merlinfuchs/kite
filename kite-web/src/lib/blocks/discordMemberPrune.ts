import { numericOrPlaceholder } from "../flow/dataSchema";
import { nodeActionMemberPruneResultSchema } from "../flow/resultSchema";
import { guildTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberPrune: BlockDefinition = {
  type: "action_member_prune",
  title: "Prune members",
  description: "Remove inactive members without roles from the server",
  icon: "users-round",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  audit_log_reason: true,
  strict_settings: true,
  fields: [
    guildTargetSetting,
    {
      name: "member_prune_days",
      schema: numericOrPlaceholder(
        "Remove members that haven't been active for this many days, between 1 and 30. Members with any role are never removed."
      ),
    },
  ],
  result: { schema: nodeActionMemberPruneResultSchema },
  run: { kind: "custom" },
};
