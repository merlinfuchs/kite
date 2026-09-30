import { numericOrPlaceholder } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlLoop: BlockDefinition = {
  type: "control_loop",
  title: "Run a loop",
  description: "Run a set of actions multiple times.",
  icon: "repeat-2",
  category: "Loops",
  outputs: [],
  owns: ["control_loop_end", "control_loop_each"],
  component: "control_loop",
  fields: [
    {
      name: "loop_count",
      schema: numericOrPlaceholder("How many times the loop runs."),
    },
  ],
  run: { kind: "custom" },
};
