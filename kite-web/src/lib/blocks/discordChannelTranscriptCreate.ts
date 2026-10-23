import { z } from "zod";
import { numericOrPlaceholder } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordChannelTranscriptCreate: BlockDefinition = {
  type: "action_channel_transcript_create",
  id_prefix: "transcript",
  title: "Create channel transcript",
  description: "Save a channel's messages as an HTML file",
  icon: "file-text",
  category: "Channels",
  requires: ["discord"],
  credits: 5,
  strict_settings: true,
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder(
        "ID of the channel or thread to make the transcript of."
      ),
    },
    {
      name: "transcript_data",
      schema: z
        .object({
          message_limit: numericOrPlaceholder(
            "How many of the most recent messages to include, between 1 and 1000. Defaults to 1000."
          ).optional(),
        })
        .optional()
        .describe("How many messages the transcript includes."),
    },
  ],
  result: {
    thing: "file",
    description:
      "The transcript as an HTML file. Put the result on its own, e.g. {{result('id')}}, in the text of a block that sends a message to attach the file to it.",
  },
  run: { kind: "custom" },
};
