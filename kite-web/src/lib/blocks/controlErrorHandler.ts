import { z } from "zod";
import { errorColor } from "./colors";
import { BlockDefinition } from "./types";

export const controlErrorHandler: BlockDefinition = {
  type: "control_error_handler",
  title: "Handle Errors",
  description: "Handle errors that occur in the flow after this block.",
  icon: "circle-alert",
  color: errorColor,
  category: "Errors",
  outputs: ["error", "default"],
  output_labels: {
    error: { label: "Handle Error", color: errorColor },
    default: { label: "Try Blocks" },
  },
  component: "branches",
  result: { schema: z.string().describe("The error message.") },
  run: { kind: "custom" },
};
