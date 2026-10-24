import { numericOrPlaceholder } from "../flow/dataSchema";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageCreate: BlockDefinition = {
  type: "action_message_create",
  id_prefix: "message",
  title: "Create channel message",
  description: "Bot sends a message to a channel",
  icon: "message-circle-plus",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder("ID of the channel to send the message to."),
    },
    {
      name: "reply_target",
      schema: numericOrPlaceholder(
        "ID of a message in the same channel to reply to. Leave unset to send a message that isn't a reply."
      ).optional(),
    },
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  result: { thing: "discord_message" },
  run: { kind: "custom" },
};
