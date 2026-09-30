import { z } from "zod";
import { numericRegex } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordOptionCommandPermissions: BlockDefinition = {
  type: "option_command_permissions",
  title: "Command Permissions",
  description:
    "Make the command only available to users with the specified permissions.",
  icon: "shield-check",
  category: "Commands",
  component: "option",
  requires: ["discord"],
  custom_label: false,
  fields: [
    {
      name: "command_permissions",
      schema: z
        .string()
        .regex(numericRegex)
        .describe(
          "Discord permission bitfield. Only members with all of these permissions can see and use the command."
        ),
    },
  ],
  run: { kind: "custom" },
};
