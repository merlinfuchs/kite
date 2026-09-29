import { nodeActionForumPostCreateDataSchema } from "../flow/dataSchema";
import { nodeActionForumPostCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordForumPostCreate: BlockDefinition = {
  type: "action_forum_post_create",
  title: "Create forum post",
  description: "Create a forum post",
  icon: "message-circle-plus",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionForumPostCreateDataSchema,
  inputs: [
    "channel_target",
    "channel_data",
    "audit_log_reason",
    "temporary_name",
    "custom_label",
  ],
  result: { schema: nodeActionForumPostCreateResultSchema },
  run: { kind: "custom" },
};
