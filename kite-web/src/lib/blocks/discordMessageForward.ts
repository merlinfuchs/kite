import { numericOrPlaceholder } from "../flow/dataSchema";
import { messageTargetSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageForward: BlockDefinition = {
  type: "action_message_forward",
  id_prefix: "forward",
  title: "Forward message",
  description: "Bot forwards a message to another channel",
  icon: "forward",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "channel_target",
      input: "forward_source_channel_target",
      schema: numericOrPlaceholder(
        "ID of the channel the message to forward is in. Defaults to the channel the flow runs in."
      ).optional(),
    },
    messageTargetSetting,
    {
      name: "forward_channel_target",
      schema: numericOrPlaceholder(
        "ID of the channel to forward the message to."
      ),
    },
  ],
  result: { thing: "discord_message", description: "The forwarded message" },
  run: { kind: "custom" },
};
