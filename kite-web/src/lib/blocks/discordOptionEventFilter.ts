import { nodeOptionEventFilterSchema } from "../flow/dataSchema";
import { BlockDefinition } from "./types";

export const discordOptionEventFilter: BlockDefinition = {
  type: "option_event_filter",
  title: "Event Filter",
  description: "Filter events based on their properties.",
  icon: "filter",
  category: "Events",
  component: "option",
  requires: ["discord"],
  schema: nodeOptionEventFilterSchema,
  inputs: ["event_filter_target", "event_filter_mode", "event_filter_value"],
  run: { kind: "custom" },
};
