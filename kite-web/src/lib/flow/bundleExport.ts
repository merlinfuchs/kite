import { FlowData } from "../types/flow.gen";
import { eventListenerLabel } from "./eventTypes";
import { bundleContentSchema, entryData, TemplateSetting } from "./templates";

export type ExportedBundle = {
  name: string;
  settings: TemplateSetting[];
  // Commands and event listeners can be left out of an import.
  items: { key: string; label: string }[];
  // Message templates and variables are always imported, the blocks use them.
  messages: { key: string; label: string }[];
  variables: { key: string; label: string }[];
  flows: FlowData[];
};

// parseExportedBundle returns several items exported from an app as one share
// code, or undefined if the data isn't that. Item ids are the ids they had in
// that app, which their blocks use for message templates and variables.
export function parseExportedBundle(data: unknown): ExportedBundle | undefined {
  const res = bundleContentSchema.safeParse(data);
  if (!res.success) return undefined;
  const bundle = res.data;

  return {
    name: bundle.name,
    settings: bundle.settings,
    items: [
      ...bundle.commands.map((c) => ({
        key: c.id,
        label: `/${String(entryData(c.flow_source).name ?? "")}`,
      })),
      ...bundle.event_listeners.map((l) => ({
        key: l.id,
        label: eventListenerLabel(
          l.source,
          String(entryData(l.flow_source).event_type ?? "")
        ),
      })),
    ],
    messages: bundle.messages.map((m) => ({ key: m.id, label: m.name })),
    variables: bundle.variables.map((v) => ({ key: v.id, label: v.name })),
    flows: [
      ...bundle.commands.map((c) => c.flow_source),
      ...bundle.event_listeners.map((l) => l.flow_source),
      ...bundle.messages.flatMap((m) => Object.values(m.flow_sources)),
    ] as FlowData[],
  };
}
