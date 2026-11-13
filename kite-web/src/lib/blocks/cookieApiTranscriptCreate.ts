import { z } from "zod";
import { flowChannelField } from "./fields";
import { BlockDefinition } from "./types";

export const cookieApiTranscriptCreate: BlockDefinition = {
  type: "action_cookie_api_transcript_create",
  id_prefix: "transcript",
  title: "Create transcript",
  description: "Save the messages of a channel as a web page with Cookie API",
  icon: "scroll-text",
  category: "Cookie API",
  strict_settings: true,
  requires: ["discord"],
  run: {
    kind: "request",
    integration: "cookie_api",
    operation: "create_transcript",
    inject: [{ in: "body", name: "bot_token", value: "discord_bot_token" }],
  },
  fields: [
    {
      ...flowChannelField,
      description:
        "ID of the channel, thread or forum post. Leave empty to use the channel the flow runs in. Up to its last 500 messages are saved.",
    },
    {
      name: "transcript_title",
      target: "title",
      label: "Title",
      description: "Title of the link's embed.",
    },
    {
      name: "transcript_description",
      target: "description",
      label: "Description",
      description: "Description of the link's embed.",
    },
    {
      name: "transcript_password",
      target: "password",
      label: "Password",
      description:
        "Password needed to open the transcript. Anyone with the link can open it without one, so set one for private channels like tickets.",
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
