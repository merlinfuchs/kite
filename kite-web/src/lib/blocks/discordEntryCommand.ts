import { nodeEntryCommandDataSchema } from "../flow/dataSchema";
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
  schema: nodeEntryCommandDataSchema,
  inputs: ["name", "description"],
  run: { kind: "custom" },
};
