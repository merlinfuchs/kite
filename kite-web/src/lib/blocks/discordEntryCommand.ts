import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordEntryCommand: BlockDefinition = {
  type: "entry_command",
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
        .regex(
          /^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$/,
          "Must be only lowercase alphanumeric characters and underscores, and have at most 3 words"
        )
        .describe(
          "Name of the slash command. Up to three space separated words create subcommands, e.g. 'ticket open'."
        ),
    },
    {
      name: "description",
      schema: z
        .string()
        .max(100)
        .min(1)
        .describe("Description of the command shown in Discord."),
    },
  ],
  run: { kind: "custom" },
};
