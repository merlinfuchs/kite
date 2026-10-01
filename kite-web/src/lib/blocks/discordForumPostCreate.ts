import { numericOrPlaceholder } from "../flow/dataSchema";
import { nodeActionForumPostCreateResultSchema } from "../flow/resultSchema";
import { channelDataSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordForumPostCreate: BlockDefinition = {
  type: "action_forum_post_create",
  title: "Create forum post",
  description: "Create a forum post",
  icon: "message-circle-plus",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder("ID of the forum channel."),
    },
    channelDataSetting,
  ],
  audit_log_reason: true,
  result: { schema: nodeActionForumPostCreateResultSchema },
  run: { kind: "custom" },
};
