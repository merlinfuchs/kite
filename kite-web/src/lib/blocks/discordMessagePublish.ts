import { channelTargetField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessagePublish: BlockDefinition = {
  type: "action_message_publish",
  id_prefix: "message",
  title: "Publish message",
  description: "Publish a message in an announcement channel",
  icon: "megaphone",
  category: "Messages",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "crosspost_message",
    method: "POST",
    path: "/channels/{channel_id}/messages/{message_id}/crosspost",
  },
  fields: [channelTargetField, messageTargetField],
  result: {
    thing: "discord_message",
    description: "The published message",
  },
};
