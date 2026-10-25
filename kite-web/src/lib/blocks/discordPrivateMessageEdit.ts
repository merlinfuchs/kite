import { numericOrPlaceholder } from "../flow/dataSchema";
import {
  messageDataField,
  messageTargetSetting,
  messageTemplateField,
  requireMessage,
} from "./fields";
import { BlockDefinition } from "./types";

export const discordPrivateMessageEdit: BlockDefinition = {
  type: "action_private_message_edit",
  id_prefix: "dm",
  title: "Edit direct message",
  description: "Bot edits a direct message it sent to a user",
  icon: "pen",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "user_target",
      schema: numericOrPlaceholder("ID of the user the message was sent to."),
    },
    messageTargetSetting,
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  result: { thing: "discord_message" },
  run: { kind: "custom" },
};
