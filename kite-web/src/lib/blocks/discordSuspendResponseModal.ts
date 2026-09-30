import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordSuspendResponseModal: BlockDefinition = {
  type: "suspend_response_modal",
  title: "Show Modal",
  description:
    "Show a modal to the user and suspend the flow until the user submits the modal.",
  icon: "picture-in-picture-2",
  category: "Responses",
  component: "suspend",
  requires: ["discord"],
  fields: [
    {
      name: "modal_data",
      type: "string",
      schema: z
        .object({
          title: templated(z.string().max(45).min(1), "Title of the modal."),
          components: z
            .array(
              z.object({
                components: z
                  .array(
                    z.object({
                      custom_id: z
                        .string()
                        .max(100)
                        .min(1)
                        .describe(
                          "Identifier of the input. The submitted value can be read with {{input('custom_id')}}. This is fixed text, placeholders aren't supported."
                        ),
                      label: templated(
                        z.string().max(45).min(1),
                        "Label shown above the input."
                      ),
                      style: z
                        .literal(1)
                        .or(z.literal(2))
                        .describe(
                          "1 for a single line input, 2 for a paragraph."
                        ),
                      required: z
                        .boolean()
                        .optional()
                        .describe("Whether the input has to be filled in."),
                      min_length: z
                        .number()
                        .optional()
                        .describe("Minimum length of the entered text."),
                      max_length: z
                        .number()
                        .optional()
                        .describe("Maximum length of the entered text."),
                      value: templated(
                        z.string().max(4000).min(1),
                        "Value the input is pre-filled with."
                      ).optional(),
                      placeholder: templated(
                        z.string().max(4000).min(1),
                        "Text shown while the input is empty."
                      ).optional(),
                    })
                  )
                  .min(1)
                  .max(1)
                  .describe("The text input of the row."),
              })
            )
            .min(1)
            .max(5)
            .describe("Rows of the modal, each holding one text input."),
        })
        .describe("The modal to show."),
    },
  ],
  run: { kind: "custom" },
};
