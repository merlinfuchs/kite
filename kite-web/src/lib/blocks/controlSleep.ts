import { decimalRegex, numericOrPlaceholder } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const controlSleep: BlockDefinition = {
  type: "control_sleep",
  title: "Wait",
  description: "Pause the flow for a set amount of time.",
  icon: "timer",
  category: "Others",
  component: "control_sleep",
  custom_label: false,
  fields: [
    {
      name: "sleep_duration_seconds",
      type: "string",
      schema: numericOrPlaceholder(
        "How many seconds to wait before continuing.",
        decimalRegex
      ),
    },
  ],
  run: { kind: "custom" },
};
