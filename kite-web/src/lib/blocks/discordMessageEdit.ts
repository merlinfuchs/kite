import {
  channelTargetSetting,
  messageDataField,
  messageTargetSetting,
  messageTemplateField,
  requireMessage,
} from "./fields";
import { nodeActionMessageEditResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordMessageEdit: BlockDefinition = {
  type: "action_message_edit",
  title: "Edit channel message",
  description: "Bot edits an existing message in a channel",
  icon: "pen",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  fields: [
    channelTargetSetting,
    messageTargetSetting,
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  result: { schema: nodeActionMessageEditResultSchema },
  run: { kind: "custom" },
};
