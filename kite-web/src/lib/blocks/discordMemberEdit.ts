import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { guildTargetSetting, userTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordMemberEdit: BlockDefinition = {
  type: "action_member_edit",
  title: "Edit member nickname",
  description: "Edit a member in the server",
  icon: "user-round-pen",
  category: "Members",
  requires: ["discord"],
  credits: 1,
  fields: [
    guildTargetSetting,
    userTargetSetting,
    {
      name: "member_data",
      schema: z
        .object({
          nick: templated(z.string(), "New nickname of the member."),
        })
        .describe("The changes to make to the member."),
    },
  ],
  audit_log_reason: true,
  run: { kind: "custom" },
};
