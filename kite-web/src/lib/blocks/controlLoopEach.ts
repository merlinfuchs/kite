import { BlockDefinition } from "./types";

export const controlLoopEach: BlockDefinition = {
  type: "control_loop_each",
  title: "Each loop iteration",
  description: "Run actions for each iteration of the loop.",
  icon: "repeat-2",
  fixed: true,
  component: "control_loop_each",
  custom_label: false,
  run: { kind: "custom" },
};
