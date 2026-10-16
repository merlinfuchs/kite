import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordOptionCommandArgument: BlockDefinition = {
  type: "option_command_argument",
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
        .optional()
        .describe(
          "Fixed choices the user picks from, for string and integer arguments. Choices with an empty name or value are ignored."
        ),
    },
  ],
  run: { kind: "custom" },
};
