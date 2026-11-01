import { z } from "zod";
import { autocompleteMaxChoices } from "./discordResponseAutocomplete";
import { BlockDefinition, SettingsSchema } from "./types";

export const discordOptionCommandArgument: BlockDefinition = {
  type: "option_command_argument",
  id_prefix: "arg",
  title: "Command Argument",
  description: "Argument for a command.",
  icon: "text-cursor-input",
  category: "Commands",
  component: "option_command_argument",
  requires: ["discord"],
  custom_label: false,
  fields: [
    {
      name: "name",
      schema: z
        .string()
        .max(32)
        .min(1)
        .regex(/^[a-z0-9_]+$/)
        .describe(
          "Name of the argument. Its value can be read with {{arg('name')}}."
        ),
    },
    {
      name: "description",
      schema: z
        .string()
        .max(100)
        .min(1)
        .regex(/\S/, "Can't be only spaces.")
        .describe("Description of the argument shown in Discord."),
    },
    {
      name: "command_argument_type",
      schema: z
        .enum([
          "string",
          "integer",
          "boolean",
          "user",
          "channel",
          "role",
          "mentionable",
          "number",
          "attachment",
        ])
        .describe("Type of value the user has to enter."),
    },
    {
      name: "command_argument_required",
      schema: z
        .boolean()
        .optional()
        .describe("Whether the user has to fill in the argument."),
    },
    {
      name: "command_argument_min_value",
      schema: z
        .number()
        .min(-Number.MAX_SAFE_INTEGER)
        .max(Number.MAX_SAFE_INTEGER)
        .optional()
        .describe("Smallest allowed value for integer and number arguments."),
    },
    {
      name: "command_argument_max_value",
      schema: z
        .number()
        .min(-Number.MAX_SAFE_INTEGER)
        .max(Number.MAX_SAFE_INTEGER)
        .optional()
        .describe("Largest allowed value for integer and number arguments."),
    },
    {
      name: "command_argument_max_length",
      schema: z
        .number()
        .int()
        .min(0)
        .max(6000)
        .optional()
        .describe("Maximum length for string arguments, up to 6000."),
    },
    {
      name: "command_argument_autocomplete",
      schema: z
        .boolean()
        .optional()
        .describe(
          "Suggest choices while the user types, for string, integer and number arguments. Without blocks after the argument's autocomplete output, the choices that contain what was typed are shown. Blocks after the autocomplete output run on every key press instead, and show choices with a Show autocomplete choices block. The user can still enter any value."
        ),
    },
    {
      name: "command_argument_choices",
      schema: z
        .array(
          z.object({
            name: z.string().max(100).describe("Name shown to the user."),
            value: z
              .string()
              .max(100)
              .describe("Value the flow receives when the choice is picked."),
          })
        )
        .max(autocompleteMaxChoices)
        .optional()
        .describe(
          "Fixed choices the user picks from, for string, integer and number arguments. Choices with an empty name or value are ignored. Up to 25, or 500 with autocomplete."
        ),
    },
  ],
  // Discord rejects more than 25 choices when the command is deployed. With
  // autocomplete, Kite shows the matching ones instead.
  refine: (schema: SettingsSchema) =>
    schema.superRefine((data, ctx) => {
      if (
        !data.command_argument_autocomplete &&
        (data.command_argument_choices?.length ?? 0) > 25
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["command_argument_choices"],
          message:
            "Discord allows up to 25 choices. Turn on autocomplete for more.",
        });
      }
    }),
  run: { kind: "custom" },
};
