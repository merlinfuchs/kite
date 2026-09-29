import { z } from "zod";
import { messageResultSchema } from "../../../flow/resultSchema";
import { IntegrationBlock } from "../../types";

export const messageList: IntegrationBlock = {
  type: "action_message_list",
  integration: "discord",
  operation: "list_messages",
  method: "GET",
  path: "/channels/{channel_id}/messages",
  title: "List channel messages",
  description: "Get the latest messages of a channel",
  icon: "messages-square",
  category: "Messages",
  credits: 1,
  fields: [
    {
      name: "channel_target",
      in: "path",
      target: "channel_id",
      type: "snowflake",
      label: "Channel",
      description:
        "ID of the channel. Leave empty to use the channel the flow runs in.",
      fallback: "channel",
    },
    {
      name: "limit",
      in: "query",
      type: "integer",
      label: "Number of Messages",
      description: "How many messages to get, from 1 to 100. Defaults to 50.",
      min: 1,
      max: 100,
    },
    {
      name: "before",
      in: "query",
      type: "snowflake",
      label: "Before Message",
      description: "Only get messages sent before the message with this ID.",
    },
    {
      name: "after",
      in: "query",
      type: "snowflake",
      label: "After Message",
      description: "Only get messages sent after the message with this ID.",
    },
  ],
  result: {
    thing: "discord_message",
    list: true,
    schema: z.array(messageResultSchema).describe("The messages, newest first"),
  },
};
