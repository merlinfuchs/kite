import { nodeControlSleepDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlSleep: BlockDefinition = {
  type: "control_sleep",
  title: "Wait",
  description: "Pause the flow for a set amount of time.",
  icon: "timer",
  category: "Others",
  component: "control_sleep",
  schema: nodeControlSleepDataSchema,
  inputs: ["sleep_duration_seconds"],
  run: { kind: "custom" },
};
