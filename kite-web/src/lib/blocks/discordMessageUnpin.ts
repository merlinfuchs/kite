import { nodeActionMessagePinDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessageUnpin: BlockDefinition = {
  type: "action_message_unpin",
  title: "Unpin channel message",
  description: "Bot unpins a message in a channel",
  icon: "pin-off",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessagePinDataSchema,
  inputs: [
    "channel_target",
    "message_target",
    "audit_log_reason",
    "custom_label",
  ],
  run: { kind: "custom" },
};
