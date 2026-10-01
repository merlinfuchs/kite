import { z } from "zod";
import { flowChannelField } from "./fields";
import { BlockDefinition } from "./types";

export const cookieApiTranscriptCreate: BlockDefinition = {
  type: "action_cookie_api_transcript_create",
  title: "Create transcript",
  description: "Save the messages of a channel as a web page with Cookie API",
  icon: "scroll-text",
  category: "Cookie API",
  credits: 3,
  strict_settings: true,
  requires: ["discord"],
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "create_transcript",
    method: "POST",
    path: "/api/transcript",
  },
  fields: [
    { name: "bot_token", in: "body", type: "discord_bot_token" },
    {
      ...flowChannelField,
      in: "query",
      description:
        "ID of the channel, thread or forum post. Leave empty to use the channel the flow runs in. Up to its last 500 messages are saved.",
    },
    {
      name: "transcript_title",
      in: "body",
      target: "title",
      type: "string",
      label: "Title",
      description: "Title of the link's embed.",
    },
    {
      name: "transcript_description",
      in: "body",
      target: "description",
      type: "string",
      label: "Description",
      description: "Description of the link's embed.",
    },
    {
      name: "transcript_password",
      in: "body",
      target: "password",
      type: "string",
      label: "Password",
      description: "Password needed to open the transcript.",
    },
  ],
  result: {
    schema: z
      .object({
        url: z.string().describe("Link to the transcript"),
      })
      .describe("The created transcript"),
  },
};
