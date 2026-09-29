import { nodeSuspendResponseModalDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordSuspendResponseModal: BlockDefinition = {
  type: "suspend_response_modal",
  title: "Show Modal",
  description:
    "Show a modal to the user and suspend the flow until the user submits the modal.",
  icon: "picture-in-picture-2",
  category: "Responses",
  component: "suspend",
  requires: ["discord"],
  schema: nodeSuspendResponseModalDataSchema,
  inputs: ["modal_data", "custom_label"],
  run: { kind: "custom" },
};
