import { z } from "zod";
import { BlockDefinition } from "./types";
import { isContextMenuCommand } from "../flow/commandTypes";

// User and Message commands can have any name and no description, slash
// commands can't. The flow AI only sees these, as it only makes slash
// commands.
const slashCommandName = z
  .string()
  .max(32)
  .min(1)
  .regex(
    /^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$/,
    "Must be only lowercase alphanumeric characters and underscores, and have at most 3 words"
  )
  .describe(
    "Name of the slash command. Up to three space separated words create subcommands, e.g. 'ticket open'."
  );

const slashCommandDescription = z
  .string()
  .max(100)
  .min(1)
  .regex(/\S/, "Can't be only spaces.")
  .describe("Description of the command shown in Discord.");

export const discordEntryCommand: BlockDefinition = {
  type: "entry_command",
  id_prefix: "command",
  title: "Command",
  description: "Command entry. Drop different actions and options here!",
  icon: "square-slash",
  contexts: ["command"],
  fixed: true,
  component: "entry_command",
  requires: ["discord"],
  custom_label: false,
  fields: [
    {
      name: "command_type",
      schema: z
        .enum(["chat_input", "user", "message"])
        .optional()
        .describe(
          "A slash command, or a User or Message command in the Apps menu of a user or message."
        ),
      hidden_from_ai: true,
    },
    {
      name: "name",
      schema: z
        .string()
        .max(32)
        .min(1)
        .regex(/\S/, "Can't be only spaces.")
        .describe("Name of the command shown in Discord."),
      ai_schema: slashCommandName,
    },
    {
      name: "description",
      schema: z
        .string()
        .max(100)
        .optional()
        .describe("Description of a slash command shown in Discord."),
      ai_schema: slashCommandDescription,
    },
    {
      name: "command_argument_order",
      schema: z
        .array(z.string())
        .max(25)
        .optional()
        .describe(
          "IDs of the command's option_command_argument blocks in the order Discord shows them. Required arguments always come before optional ones, and arguments missing here come last."
        ),
      hidden_from_ai: true,
    },
  ],
  refine: (schema) =>
    schema.superRefine((data, ctx) => {
      if (isContextMenuCommand(data)) return;
      for (const [field, fieldSchema] of [
        ["name", slashCommandName],
        ["description", slashCommandDescription],
      ] as const) {
        const res = fieldSchema.safeParse(data[field]);
        for (const issue of res.error?.issues ?? []) {
          ctx.addIssue({ ...issue, path: [field, ...issue.path] });
        }
      }
    }),
  run: { kind: "custom" },
};
