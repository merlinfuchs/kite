import { channelDataSchema, numericOrPlaceholder } from "../flow/dataSchema";
import { nodeActionForumPostCreateResultSchema } from "../flow/resultSchema";
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
      type: "snowflake",
      schema: numericOrPlaceholder("ID of the forum channel."),
    },
    {
      name: "channel_data",
      type: "string",
      schema: channelDataSchema,
    },
  ],
  audit_log_reason: true,
  result: { schema: nodeActionForumPostCreateResultSchema },
  run: { kind: "custom" },
};
