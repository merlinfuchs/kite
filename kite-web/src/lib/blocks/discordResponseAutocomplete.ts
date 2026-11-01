import { z } from "zod";
import { templated } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

// Limits of the service, see autocomplete.go in pkg/flow. Discord shows the
// first 25 choices.
export const autocompleteMaxChoices = 500;

export const discordResponseAutocomplete: BlockDefinition = {
  type: "action_response_autocomplete",
  id_prefix: "autocomplete",
  title: "Show autocomplete choices",
  description:
    "Shows the choices for what the user typed in an argument with autocomplete",
  icon: "list-filter",
  category: "Responses",
  contexts: ["command"],
  needs_interaction: true,
  requires: ["discord"],
  credits: 1,
  strict_settings: true,
  fields: [
    {
      name: "autocomplete_choices",
      schema: z
        .array(
          z.object({
            name: templated(
              z.string().max(100),
              "Name of the choice shown to the user."
            ),
            value: templated(
              z.string().max(100).optional(),
              "Value the command receives when the choice is picked. The name when empty."
            ),
          })
        )
        .max(autocompleteMaxChoices)
        .optional()
        .describe(
          "Choices to show. Choices with an empty name, and values that don't fit an integer or number argument, are left out."
        ),
    },
    {
      name: "autocomplete_choices_list",
      schema: templated(
        z.string().max(2000).optional(),
        "A list to show more choices from, like {{var('fruits')}}: texts, or objects with a name and value. Text has a choice on each line."
      ),
    },
    {
      name: "autocomplete_filter",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Only show the choices that contain what the user typed, those that start with it first."
        ),
    },
  ],
  run: { kind: "custom" },
};
