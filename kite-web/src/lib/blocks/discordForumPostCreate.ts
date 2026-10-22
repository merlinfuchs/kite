import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
import { BlockDefinition } from "./types";

export const discordForumPostCreate: BlockDefinition = {
  type: "action_forum_post_create",
  title: "Create forum post",
  description: "Create a post in a forum or media channel",
  icon: "message-circle-plus",
  category: "Channels",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder(
        "ID of the forum or media channel to create the post in."
      ),
    },
    {
      name: "channel_data",
      input: "forum_post_data",
      schema: z
        .object({
          name: templated(z.string().max(100).min(1), "Title of the post."),
          applied_tags: templated(
            z.string(),
            "Names or IDs of up to 5 of the forum's tags, separated by commas."
          ).optional(),
        })
        .describe("Title and tags of the post."),
    },
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  audit_log_reason: true,
  result: { thing: "discord_channel" },
  run: { kind: "custom" },
};
