import { nodeEntryEventDataSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordEntryEvent: BlockDefinition = {
  type: "entry_event",
  title: "Listen for Event",
  description:
    "Listens for an event to trigger the flow. Drop different actions here!",
  icon: "satellite-dish",
  contexts: ["event_discord", "event_schedule"],
  fixed: true,
  component: "entry_event",
  requires: ["discord"],
  schema: nodeEntryEventDataSchema,
  inputs: ["event_type", "event_schedule_cron", "description"],
  run: { kind: "custom" },
};
