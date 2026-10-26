import { BlockDefinition } from "./types";

export const controlLoopExit: BlockDefinition = {
  type: "control_loop_exit",
  id_prefix: "loop",
  title: "Exit loop",
  description: "Exit out of the loop.",
  icon: "log-out",
  category: "Loops",
  outputs: [],
  component: "terminal",
  custom_label: false,
  run: { kind: "custom" },
};
