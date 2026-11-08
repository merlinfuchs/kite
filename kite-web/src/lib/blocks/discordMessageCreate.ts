import { PermissionFlag } from "../discord/permissions";
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
  permissions: (data) => {
    const perms: PermissionFlag[] = ["VIEW_CHANNEL", "SEND_MESSAGES"];
    const msg = data?.message_data as any;
    if (Array.isArray(msg?.embeds) && msg.embeds.length > 0) {
      perms.push("EMBED_LINKS");
    }
    if (Array.isArray(msg?.attachments) && msg.attachments.length > 0) {
      perms.push("ATTACH_FILES");
    }
    return perms;
  },
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
