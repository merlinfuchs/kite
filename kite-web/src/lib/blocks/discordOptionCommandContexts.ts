import { nodeOptionCommandContextsSchema } from "../flow/dataSchema";
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
  schema: nodeOptionCommandContextsSchema,
  inputs: ["command_contexts", "command_integrations"],
  run: { kind: "custom" },
};
