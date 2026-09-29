import { nodeOptionCommandPermissionsSchema } from "../flow/dataSchema";
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
  schema: nodeOptionCommandPermissionsSchema,
  inputs: ["command_permissions"],
  run: { kind: "custom" },
};
