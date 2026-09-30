import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordResponseDefer: BlockDefinition = {
  type: "action_response_defer",
  title: "Defer response",
  description:
    "Bot defers the response to the interaction to give time for further processing",
  icon: "message-circle-question",
  category: "Responses",
  requires: ["discord"],
  credits: 1,
  fields: [
    {
      name: "message_ephemeral",
      type: "boolean",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Whether only the user who triggered the flow can see the response that follows."
        ),
    },
  ],
  run: { kind: "custom" },
};
