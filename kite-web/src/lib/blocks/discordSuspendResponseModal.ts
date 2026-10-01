import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { normalizeModalData } from "../flow/modal";
import { BlockDefinition } from "./types";

const modalCustomIdSchema = z
  .string()
  .max(100)
  .min(1)
  .describe(
    "Identifier of the input. The submitted value can be read with {{input('custom_id')}}. This is fixed text, placeholders aren't supported."
  );

const modalRequiredSchema = z
  .boolean()
  .optional()
  .describe("Whether the input has to be filled in.");

const modalPlaceholderSchema = templated(
  z.string().max(150).min(1),
  "Text shown while nothing is entered or picked."
).optional();

const modalMinValuesSchema = (max: number) =>
  z
    .number()
    .int()
    .min(0)
    .max(max)
    .optional()
    .describe("Minimum number of options that have to be picked.");

const modalMaxValuesSchema = (max: number) =>
  z
    .number()
    .int()
    .min(1)
    .max(max)
    .optional()
    .describe("Maximum number of options that can be picked. Defaults to 1.");

const modalOptionsSchema = (min: number, max: number) =>
  z
    .array(
      z.object({
        label: templated(
          z.string().max(100).min(1),
          "Label of the option shown to the user."
        ),
        value: templated(
          z.string().max(100),
          "Value input() returns when the option is picked. Defaults to the label."
        ).optional(),
        description: templated(
          z.string().max(100),
          "Smaller text shown below the label."
        ).optional(),
        default: z
          .boolean()
          .optional()
          .describe("Whether the option is picked when the modal opens."),
      })
    )
    .min(min)
    .max(max)
    .describe("Options to pick from. Their values must be unique.");

const modalEntityNames = {
  user_select: "members",
  role_select: "roles",
  mentionable_select: "members and roles",
} as const;

function modalEntitySelectSchema<T extends keyof typeof modalEntityNames>(
  type: T
) {
  return z.object({
    type: z
      .literal(type)
      .describe(`A select menu of the server's ${modalEntityNames[type]}.`),
    custom_id: modalCustomIdSchema,
    required: modalRequiredSchema,
    placeholder: modalPlaceholderSchema,
    min_values: modalMinValuesSchema(25),
    max_values: modalMaxValuesSchema(25),
  });
}

const modalInputSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("text_input").describe("A text box."),
    custom_id: modalCustomIdSchema,
    style: z
      .literal(1)
      .or(z.literal(2))
      .describe("1 for a single line input, 2 for a paragraph."),
    required: modalRequiredSchema,
    min_length: z
      .number()
      .int()
      .min(0)
      .max(4000)
      .optional()
      .describe("Minimum length of the entered text."),
    max_length: z
      .number()
      .int()
      .min(1)
      .max(4000)
      .optional()
      .describe("Maximum length of the entered text."),
    value: templated(
      z.string().max(4000).min(1),
      "Value the input is pre-filled with."
    ).optional(),
    placeholder: modalPlaceholderSchema,
  }),
  z.object({
    type: z
      .literal("string_select")
      .describe("A select menu with your own options."),
    custom_id: modalCustomIdSchema,
    required: modalRequiredSchema,
    placeholder: modalPlaceholderSchema,
    min_values: modalMinValuesSchema(25),
    max_values: modalMaxValuesSchema(25),
    options: modalOptionsSchema(1, 25),
  }),
  modalEntitySelectSchema("user_select"),
  modalEntitySelectSchema("role_select"),
  modalEntitySelectSchema("mentionable_select"),
  z.object({
    type: z
      .literal("channel_select")
      .describe("A select menu of the server's channels."),
    custom_id: modalCustomIdSchema,
    required: modalRequiredSchema,
    placeholder: modalPlaceholderSchema,
    min_values: modalMinValuesSchema(25),
    max_values: modalMaxValuesSchema(25),
    channel_types: z
      .array(z.number().int())
      .optional()
      .describe(
        "Discord channel types that can be picked, e.g. 0 for text channels. Empty allows all."
      ),
  }),
  z.object({
    type: z
      .literal("radio_group")
      .describe("A list of options of which exactly one can be picked."),
    custom_id: modalCustomIdSchema,
    required: modalRequiredSchema,
    options: modalOptionsSchema(2, 10),
  }),
  z.object({
    type: z
      .literal("checkbox_group")
      .describe("A list of options of which several can be picked."),
    custom_id: modalCustomIdSchema,
    required: modalRequiredSchema,
    min_values: modalMinValuesSchema(10),
    max_values: modalMaxValuesSchema(10),
    options: modalOptionsSchema(1, 10),
  }),
  z.object({
    type: z.literal("checkbox").describe("A single checkbox."),
    custom_id: modalCustomIdSchema,
    default: z
      .boolean()
      .optional()
      .describe(
        "Whether the checkbox starts checked. input() returns 'true' or 'false'."
      ),
  }),
]);

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
      schema: z
        .preprocess(
          (v) => (v && typeof v === "object" ? normalizeModalData(v) : v),
          z.object({
            title: templated(z.string().max(45).min(1), "Title of the modal."),
            components: z
              .array(
                z.discriminatedUnion("type", [
                  z.object({
                    type: z
                      .literal("label")
                      .describe("A label with one input below it."),
                    label: templated(
                      z.string().max(45).min(1),
                      "Label shown above the input."
                    ),
                    description: templated(
                      z.string().max(100),
                      "Smaller text shown below the label."
                    ).optional(),
                    components: z
                      .array(modalInputSchema)
                      .min(1)
                      .max(1)
                      .describe("The input the label describes."),
                  }),
                  z.object({
                    type: z
                      .literal("text_display")
                      .describe("Markdown text shown in the modal."),
                    content: templated(
                      z.string().max(4000).min(1),
                      "Markdown text shown in the modal."
                    ),
                  }),
                ])
              )
              .min(1)
              .max(5)
              .refine((c) => c.some((c) => c.type === "label"), {
                message: "A modal needs at least one input.",
              })
              .describe(
                "Components of the modal: labels that each hold one input, and text displays."
              ),
          })
        )
        .describe("The modal to show."),
    },
  ],
  run: { kind: "custom" },
};
