import { BlockDefinition } from "./types";

export const discordTypingTrigger: BlockDefinition = {
  type: "action_typing_trigger",
  title: "Show typing",
  description: "Bot shows as typing in a channel",
  icon: "message-circle-more",
  category: "Bot",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "trigger_typing_indicator",
    method: "POST",
    path: "/channels/{channel_id}/typing",
  },
  fields: [
    {
      name: "channel_target",
      in: "path",
      target: "channel_id",
      type: "snowflake",
      label: "Channel",
      description:
        "ID of the channel the bot shows as typing in. Leave empty to use the channel the flow runs in.",
      fallback: "channel",
    },
  ],
};
