import { numericOrPlaceholder } from "../flow/dataSchema";
import { nodeActionThreadCreateResultSchema } from "../flow/resultSchema";
import { channelDataSetting } from "./fields";
import { BlockDefinition } from "./types";

export const discordThreadCreate: BlockDefinition = {
  type: "action_thread_create",
  title: "Create thread",
  description: "Create a thread",
  icon: "message-circle-plus",
  category: "Channels",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "message_target",
      input: "thread_data",
      schema: numericOrPlaceholder(
        "ID of the message to start the thread from. Leave unset for a thread without a starter message."
      ).optional(),
    },
    { ...channelDataSetting, input: "thread_data" },
  ],
  audit_log_reason: true,
  result: { schema: nodeActionThreadCreateResultSchema },
  run: { kind: "custom" },
};
