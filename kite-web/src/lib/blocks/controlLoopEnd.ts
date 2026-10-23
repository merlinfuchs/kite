import { BlockDefinition } from "./types";

export const controlLoopEnd: BlockDefinition = {
  type: "control_loop_end",
  id_prefix: "loop",
  title: "After loop",
  description: "Run actions after the loop has finished.",
  icon: "corner-down-right",
  fixed: true,
  component: "control_loop_end",
  custom_label: false,
  run: { kind: "custom" },
};
