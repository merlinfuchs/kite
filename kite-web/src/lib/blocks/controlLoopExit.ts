import { BlockDefinition } from "./types";

export const controlLoopExit: BlockDefinition = {
  type: "control_loop_exit",
  title: "Exit loop",
  description: "Exit out of the loop.",
  icon: "log-out",
  category: "Loops",
  outputs: [],
  component: "control_loop_exit",
  custom_label: false,
  run: { kind: "custom" },
};
