import { nodeOptionCommandArgumentDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordOptionCommandArgument: BlockDefinition = {
  type: "option_command_argument",
  title: "Command Argument",
  description: "Argument for a command.",
  icon: "text-cursor-input",
  category: "Commands",
  component: "option_command_argument",
  requires: ["discord"],
  schema: nodeOptionCommandArgumentDataSchema,
  inputs: [
    "name",
    "description",
    "command_argument_type",
    "command_argument_required",
    "command_argument_min_value",
    "command_argument_max_value",
    "command_argument_max_length",
    "command_argument_choices",
  ],
  run: { kind: "custom" },
};
