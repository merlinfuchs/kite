import { nodeActionMessagePinDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessagePin: BlockDefinition = {
  type: "action_message_pin",
  title: "Pin channel message",
  description: "Bot pins a message in a channel",
  icon: "pin",
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
