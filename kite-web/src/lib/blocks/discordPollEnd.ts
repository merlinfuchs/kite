import { flowChannelField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordPollEnd: BlockDefinition = {
  type: "action_poll_end",
  id_prefix: "poll",
  title: "End poll",
  description: "Bot immediately ends a poll it sent",
  icon: "vote",
  category: "Messages",
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "poll_expire",
  },
  fields: [flowChannelField, messageTargetField],
  result: {
    thing: "discord_message",
    description: "The updated poll message",
  },
};
