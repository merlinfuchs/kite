import { z } from "zod";
import { BlockDefinition, SettingsSchema } from "./types";
import { isSlashCommand } from "../flow/commandType";

const slashCommandNameRe = /^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$/;

// Only slash commands have subcommands and a description, so their name and
// description rules don't apply to user and message commands.
function refineCommandType(schema: SettingsSchema) {
  return schema
    .superRefine((data, ctx) => {
      if (!isSlashCommand(data.command_type)) return;

      if (
        typeof data.name === "string" &&
        !slashCommandNameRe.test(data.name)
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["name"],
          message:
            "Must be only lowercase alphanumeric characters and underscores, and have at most 3 words",
        });
      }

      if (typeof data.description !== "string" || data.description === "") {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["description"],
          message: "Required",
        });
      } else if (!/\S/.test(data.description)) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["description"],
          message: "Can't be only spaces.",
        });
      }
    })
    .describe(
      "Slash commands (command_type chat_input or unset) need a description, and their name must match ^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$. User and message commands have only a name, which can have capitals and spaces."
    );
}

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
          "Type of command: chat_input for a slash command, or user or message for a command in the context menu of a user or message. Defaults to chat_input."
        ),
    },
    {
      name: "name",
      schema: z
        .string()
        .max(32)
        .min(1)
        .describe(
          "Name of the command. For slash commands, up to three space separated words of lowercase letters, numbers, dashes and underscores create subcommands, e.g. 'ticket open'."
        ),
    },
    {
      name: "description",
      schema: z
        .string()
        .max(100)
        .optional()
        .describe(
          "Description of the command shown in Discord. Required for slash commands, unused for user and message commands."
        ),
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
  refine: refineCommandType,
  run: { kind: "custom" },
};
