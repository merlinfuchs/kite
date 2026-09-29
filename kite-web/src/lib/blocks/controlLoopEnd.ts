import { nodeEmptyDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlLoopEnd: BlockDefinition = {
  type: "control_loop_end",
  title: "After loop",
  description: "Run actions after the loop has finished.",
  icon: "corner-down-right",
  fixed: true,
  component: "control_loop_end",
  schema: nodeEmptyDataSchema,
  inputs: [],
  run: { kind: "custom" },
};
