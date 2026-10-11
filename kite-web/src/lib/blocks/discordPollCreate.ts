import { z } from "zod";
import {
  emojiDataSchema,
  numericOrPlaceholder,
  templated,
} from "../flow/dataSchema";
import { nodeActionPollCreateResultSchema } from "../flow/resultSchema";
import { BlockDefinition } from "./types";

export const discordPollCreate: BlockDefinition = {
  type: "action_poll_create",
  title: "Create poll",
  description: "Bot sends a poll to a channel",
  icon: "vote",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "channel_target",
      schema: numericOrPlaceholder("ID of the channel to send the poll to."),
    },
    {
      name: "poll_data",
      schema: z
        .object({
          question: templated(
            z.string().max(300).min(1),
            "Question shown at the top of the poll."
          ),
          answers: z
            .array(
              z.object({
                text: templated(
                  z.string().max(55),
                  "Text of the answer. Answers that are empty after placeholders are filled in are skipped."
                ),
                emoji: emojiDataSchema
                  .optional()
                  .describe("Emoji shown next to the answer."),
              })
            )
            .min(1)
            .max(10)
            .describe("Answers people can vote for."),
          duration_hours: numericOrPlaceholder(
            "How many hours the poll is open for, between 1 and 768. Defaults to 24."
          ).optional(),
          allow_multiselect: z
            .boolean()
            .optional()
            .describe("Whether people can vote for more than one answer."),
        })
        .describe("The poll to send."),
    },
  ],
  result: { schema: nodeActionPollCreateResultSchema },
  run: { kind: "custom" },
};
