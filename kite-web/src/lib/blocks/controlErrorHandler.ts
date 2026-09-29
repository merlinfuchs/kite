import { nodeControlErrorHandlerDataSchema } from "../flow/dataSchema";
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
  component: "control_error_handler",
  schema: nodeControlErrorHandlerDataSchema,
  inputs: ["temporary_name", "custom_label"],
  run: { kind: "custom" },
};
