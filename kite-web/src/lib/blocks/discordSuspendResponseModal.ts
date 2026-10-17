import { z } from "zod";
import { templated } from "../flow/dataSchema";
import {
  modalEntitySelectTypes,
  modalMaxComponents,
  modalOptionCounts,
  modalOptionInputTypes,
  normalizeModalData,
} from "../flow/modal";
import { BlockDefinition } from "./types";

// The flow AI gets every field of every input type, so the input types share
// fields where they can and the descriptions are short.
const modalInputBase = z.object({
  custom_id: z
    .string()
    .max(100)
    .min(1)
    .describe(
      "Fixed identifier, no placeholders. Read the answer with {{input('custom_id')}}."
    ),
  required: z.boolean().optional().describe("Whether it must be answered."),
});

const modalSelectFields = {
  placeholder: templated(
    z.string().max(150).min(1),
    "Shown while nothing is picked. Not for radio_group and checkbox_group."
  ).optional(),
  // Discord rejects 0 for required inputs, and the service sends 0 for
  // optional ones.
  min_values: z
    .number()
    .int()
    .min(1)
    .max(25)
    .optional()
    .describe("Minimum picks."),
  max_values: z
    .number()
    .int()
    .min(1)
    .max(25)
    .optional()
    .describe(
      "Maximum picks. Defaults to 1 for selects and to all options for checkbox_group. input() returns the first pick, inputs() the list of all."
    ),
};

const modalInputSchema = z
  .discriminatedUnion("type", [
    modalInputBase.extend({
      type: z.literal("text_input").describe("A text box."),
      style: z
        .literal(1)
        .or(z.literal(2))
        .describe("1 for a single line, 2 for a paragraph."),
      min_length: z
        .number()
        .int()
        .min(0)
        .max(4000)
        .optional()
        .describe("Minimum text length."),
      max_length: z
        .number()
        .int()
        .min(1)
        .max(4000)
        .optional()
        .describe("Maximum text length."),
      value: templated(
        z.string().max(4000).min(1),
        "Pre-filled text."
      ).optional(),
      placeholder: templated(
        z.string().max(100).min(1),
        "Shown while empty."
      ).optional(),
    }),
    modalInputBase.extend({
      type: z
        .enum(modalOptionInputTypes)
        .describe(
          "string_select: a select menu with 1-25 options. radio_group: pick exactly one of 2-10 options. checkbox_group: pick any of 1-10 options."
        ),
      ...modalSelectFields,
      options: z
        .array(
          z.object({
            label: templated(z.string().max(100).min(1), "Shown to the user."),
            value: templated(
              z.string().max(100),
              "What input() returns. Defaults to the label. Must be unique."
            ).optional(),
            description: templated(
              z.string().max(100),
              "Smaller text below the label."
            ).optional(),
            default: z
              .boolean()
              .optional()
              .describe("Picked when the modal opens."),
          })
        )
        .min(1)
        .max(25)
        .describe("Options to pick from."),
    }),
    modalInputBase.extend({
      type: z
        .enum(modalEntitySelectTypes)
        .describe(
          "A select menu of the server's members, roles, both, or channels. input() returns an ID."
        ),
      ...modalSelectFields,
      channel_types: z
        .array(z.number().int())
        .optional()
        .describe(
          "channel_select only. Channel types that can be picked, e.g. 0 for text. Empty allows all."
        ),
    }),
    z.object({
      type: z
        .literal("checkbox")
        .describe("A single checkbox. input() returns 'true' or 'false'."),
      custom_id: modalInputBase.shape.custom_id,
      default: z.boolean().optional().describe("Starts checked."),
    }),
  ])
  .superRefine((input, ctx) => {
    const issue = (path: (string | number)[], message: string) =>
      ctx.addIssue({ code: z.ZodIssueCode.custom, path, message });

    if (input.type === "text_input") {
      const max = input.max_length ?? 4000;
      if ((input.min_length ?? 0) > max) {
        issue(["min_length"], `Can't be more than the maximum of ${max}.`);
      }
      return;
    }
    if (input.type === "checkbox") return;

    if (input.type === "radio_group" || input.type === "checkbox_group") {
      if (input.placeholder) {
        issue(["placeholder"], `A ${input.type} has no placeholder.`);
      }
    }
    if (input.type === "radio_group") {
      if (input.min_values !== undefined || input.max_values !== undefined) {
        issue(["max_values"], "Exactly one option of a radio_group is picked.");
      }
    }
    if ("channel_types" in input && input.channel_types?.length) {
      if (input.type !== "channel_select") {
        issue(["channel_types"], "Only a channel_select has channel types.");
      }
    }

    const limit = input.type === "checkbox_group" ? 10 : 25;
    if ((input.max_values ?? 0) > limit) {
      issue(["max_values"], `Must be at most ${limit}.`);
    }

    let max = input.max_values ?? 1;
    if ("options" in input) {
      const { options } = input;
      const [minOptions, maxOptions] = modalOptionCounts[input.type];
      if (options.length < minOptions || options.length > maxOptions) {
        issue(
          ["options"],
          `A ${input.type} needs ${minOptions} to ${maxOptions} options.`
        );
      }

      const values = new Set<string>();
      options.forEach((o, i) => {
        // An option without a value submits its label.
        const value = o.value || o.label;
        if (values.has(value)) {
          issue(
            ["options", i, "value"],
            `Another option already has the value '${value}'.`
          );
        }
        values.add(value);
      });

      if (input.max_values !== undefined && input.max_values > options.length) {
        issue(
          ["max_values"],
          `Can't be more than the ${options.length} options.`
        );
      }

      if (input.type === "radio_group") max = 1;
      if (input.type === "checkbox_group") {
        max = input.max_values ?? options.length;
      }
      if (options.filter((o) => o.default).length > max) {
        issue(
          ["options"],
          max === 1
            ? "Only one option can be picked by default."
            : `At most ${max} options can be picked by default.`
        );
      }
    }

    if ((input.min_values ?? 0) > max) {
      issue(["min_values"], `Can't be more than the maximum of ${max}.`);
    }
  });

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
                      .describe("A label with one input."),
                    label: templated(
                      z.string().max(45).min(1),
                      "Shown above the input."
                    ),
                    description: templated(
                      z.string().max(100),
                      "Smaller text below the label."
                    ).optional(),
                    components: z
                      .array(modalInputSchema)
                      .min(1)
                      .max(1)
                      .describe("The one input."),
                  }),
                  z.object({
                    type: z.literal("text_display").describe("Markdown text."),
                    content: templated(
                      z.string().max(4000).min(1),
                      "The markdown."
                    ),
                  }),
                ])
              )
              .min(1)
              .max(modalMaxComponents)
              .superRefine((components, ctx) => {
                if (!components.some((c) => c.type === "label")) {
                  ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: "A modal needs at least one input.",
                  });
                }

                const ids = new Set<string>();
                components.forEach((c, i) => {
                  if (c.type !== "label") return;
                  const id = c.components[0]?.custom_id;
                  if (ids.has(id)) {
                    ctx.addIssue({
                      code: z.ZodIssueCode.custom,
                      path: [i, "components", 0, "custom_id"],
                      message: `Another input already has the identifier '${id}'.`,
                    });
                  }
                  ids.add(id);
                });
              })
              .describe(
                "Up to 5 labels and text displays, one label at least."
              ),
          })
        )
        .describe("The modal to show."),
    },
  ],
  run: { kind: "custom" },
};
