import { BlockDefinition } from "./types";

export const controlParallel: BlockDefinition = {
  type: "control_parallel",
  title: "Run in Parallel",
  description:
    "Run the connected blocks at the same time, then continue once all of them finished.",
  icon: "split",
  category: "Others",
  outputs: ["default", "after"],
  output_labels: {
    default: { label: "Run at the Same Time" },
    after: { label: "After All Finished" },
  },
  component: "branches",
  strict_settings: true,
  run: { kind: "custom" },
};
