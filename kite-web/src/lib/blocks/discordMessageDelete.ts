import { nodeActionMessageDeleteDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordMessageDelete: BlockDefinition = {
  type: "action_message_delete",
  title: "Delete channel message",
  description: "Bot deletes an existing message in a channel",
  icon: "message-circle-x",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessageDeleteDataSchema,
  inputs: [
    "channel_target",
    "message_target",
    "audit_log_reason",
    "custom_label",
  ],
  run: { kind: "custom" },
};
