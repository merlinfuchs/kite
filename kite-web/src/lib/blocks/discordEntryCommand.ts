import { z } from "zod";
import { BlockDefinition } from "./types";

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
      name: "name",
      schema: z
        .string()
        .max(32)
        .min(1)
        .describe(
          "Name of a slash command (lowercase letters, numbers, hyphens and underscores; up to three words) or the text shown for a context menu command."
        ),
    },
    {
      name: "description",
      schema: z
        .string()
        .max(100)
        .optional()
        .describe(
          "Description shown in Discord. Required for slash commands; unused by context menu commands."
        ),
    },
    {
      name: "command_type",
      input: "command_type",
      schema: z
        .enum(["chat_input", "user", "message"])
        .optional()
        .describe(
          "Slash command, User context menu command, or Message context menu command."
        ),
      hidden_from_ai: true,
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
      const isContextual =
        data.command_type === "user" || data.command_type === "message";
      if (typeof data.name === "string") {
        const validName = isContextual
          ? data.name.trim().length > 0
          : /^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$/.test(data.name);
        if (!validName) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: ["name"],
            message: isContextual
              ? "Can't be only spaces."
              : "Must be only lowercase alphanumeric characters and underscores, and have at most 3 words",
          });
        }
      }

      if (
        !isContextual &&
        (typeof data.description !== "string" || !data.description.trim())
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["description"],
          message:
            typeof data.description === "string"
              ? "Can't be only spaces."
              : "A slash command needs a description.",
        });
      }
    }),
  run: { kind: "custom" },
};
