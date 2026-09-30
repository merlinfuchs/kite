import { z } from "zod";
import {
  guildTargetSchema,
  templated,
  userTargetSchema,
} from "../flow/dataSchema";
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
    {
      name: "guild_target",
      type: "snowflake",
      schema: guildTargetSchema.optional(),
    },
    {
      name: "user_target",
      type: "snowflake",
      schema: userTargetSchema,
    },
    {
      name: "member_data",
      type: "string",
      input: "member_nick",
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
