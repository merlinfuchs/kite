import { channelTargetField, messageTargetField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessagePin: BlockDefinition = {
  type: "action_message_pin",
  id_prefix: "message",
  title: "Pin channel message",
  description: "Bot pins a message in a channel",
  icon: "pin",
  category: "Messages",
  permissions: ["PIN_MESSAGES"],
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "deprecated_create_pin",
  },
  fields: [channelTargetField, messageTargetField],
};
