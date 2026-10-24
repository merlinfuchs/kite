import { BlockDefinition } from "./types";

export const controlFlowEnd: BlockDefinition = {
  type: "control_flow_end",
  id_prefix: "end",
  title: "End flow",
  description:
    "Stop the flow here. No blocks after it run, including branches that haven't run yet.",
  icon: "octagon-x",
  category: "Others",
  outputs: [],
  component: "control_loop_exit",
  custom_label: false,
  strict_settings: true,
  run: { kind: "custom" },
};
