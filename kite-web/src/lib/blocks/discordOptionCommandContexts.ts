import { z } from "zod";
import { BlockDefinition } from "./types";

export const discordOptionCommandContexts: BlockDefinition = {
  type: "option_command_contexts",
  title: "Command Contexts",
  description:
    "Define where your command should be available. By default, it will be available everywhere.",
  icon: "map-pin",
  category: "Commands",
  component: "option",
  requires: ["discord"],
  custom_label: false,
  fields: [
    {
      name: "command_disabled_contexts",
      schema: z
        .array(z.enum(["guild", "bot_dm", "private_channel"]))
        .optional()
        .describe(
          "Places where the command can't be used: servers, DMs with the bot, or other DMs and group DMs."
        ),
    },
    {
      name: "command_disabled_integrations",
      schema: z
        .array(z.enum(["guild_install", "user_install"]))
        .optional()
        .describe(
          "Install types the command isn't available for: installed to a server, or installed to a user."
        ),
    },
  ],
  run: { kind: "custom" },
};
