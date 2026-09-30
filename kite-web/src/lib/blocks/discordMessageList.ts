import { z } from "zod";
import { messageResultSchema } from "../flow/resultSchema";
import { flowChannelField } from "./fields";
import { BlockDefinition } from "./types";

export const discordMessageList: BlockDefinition = {
  type: "action_message_list",
  title: "List channel messages",
  description: "Get the latest messages of a channel",
  icon: "messages-square",
  category: "Messages",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "list_messages",
    method: "GET",
    path: "/channels/{channel_id}/messages",
  },
  fields: [
    flowChannelField,
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
