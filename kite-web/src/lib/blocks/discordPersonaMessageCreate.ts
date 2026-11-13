import { z } from "zod";
import { numericOrPlaceholder, templated } from "../flow/dataSchema";
import {
  messageDataField,
  messageTemplateField,
  requireMessage,
} from "./fields";
import { BlockDefinition } from "./types";

export const discordPersonaMessageCreate: BlockDefinition = {
  type: "action_persona_message_create",
  id_prefix: "persona",
  title: "Send message as",
  description: "Sends a message with another name and avatar",
  icon: "venetian-mask",
  category: "Messages",
  component: "action_message",
  requires: ["discord"],
  premium_feature: "personas",
  credits: 1,
  strict_settings: true,
  // The message is sent through a webhook the bot makes in the channel.
  permissions: () => ["VIEW_CHANNEL", "MANAGE_WEBHOOKS"],
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder(
        "ID of the channel or thread to send the message to."
      ),
    },
    {
      name: "persona_name",
      schema: templated(
        z.string().min(1).max(80),
        "Name the message is sent with, like {{user.display_name}}. It can't contain 'discord' or 'clyde'."
      ),
    },
    {
      name: "persona_avatar_url",
      schema: templated(
        z.string(),
        "URL of the avatar the message is sent with, like {{user.avatar_url}}. Leave unset for Discord's default avatar."
      ).optional(),
    },
    messageTemplateField,
    messageDataField,
  ],
  refine: requireMessage,
  result: { thing: "discord_message" },
  run: { kind: "custom" },
};
