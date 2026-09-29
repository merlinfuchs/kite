import { nodeActionStatusSetDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordStatusSet: BlockDefinition = {
  type: "action_status_set",
  title: "Set status",
  description: "Change the status and activity of the bot",
  icon: "activity",
  category: "Bot",
  requires: ["discord"],
  premium_feature: "rotating_status",
  credits: 1,
  schema: nodeActionStatusSetDataSchema,
  inputs: ["status_data", "custom_label"],
  run: { kind: "custom" },
};
