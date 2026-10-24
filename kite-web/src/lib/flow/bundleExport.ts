import { z } from "zod";
import { FlowData } from "../types/flow.gen";
import { capitalize, discordEventType, eventSourceName } from "./eventTypes";
import { templateSettingSchema } from "./templates";

// Only what the import shows is checked, the service checks the rest.
const flowSourceSchema = z.object({
  nodes: z.array(
    z.object({ id: z.string(), type: z.string(), data: z.record(z.unknown()) })
  ),
  edges: z.array(z.unknown()),
});

// exportedBundleSchema is several items exported from an app as one share
// code, in the format of the service's bundles. Item ids are the ids they had
// in that app, which their blocks use for message templates and variables.
const exportedBundleSchema = z.object({
  name: z.string(),
  settings: z.array(templateSettingSchema),
  commands: z.array(
    z.object({ id: z.string(), flow_source: flowSourceSchema })
  ),
  event_listeners: z.array(
    z.object({
      id: z.string(),
      source: z.string(),
      flow_source: flowSourceSchema,
    })
  ),
  messages: z.array(
    z.object({
      id: z.string(),
      name: z.string(),
      flow_sources: z.record(flowSourceSchema),
    })
  ),
  variables: z.array(z.object({ id: z.string(), name: z.string() })),
});

export type ExportedBundle = {
  name: string;
  settings: z.infer<typeof templateSettingSchema>[];
  // Commands and event listeners can be left out of an import.
  items: { key: string; label: string }[];
  // Message templates and variables are always imported, the blocks use them.
  messages: { key: string; label: string }[];
  variables: { key: string; label: string }[];
  flows: FlowData[];
};

// parseExportedBundle returns the exported bundle, or undefined if the data
// isn't one.
export function parseExportedBundle(data: unknown): ExportedBundle | undefined {
  const res = exportedBundleSchema.safeParse(data);
  if (!res.success) return undefined;
  const bundle = res.data;

  const entry = (flow: z.infer<typeof flowSourceSchema>) =>
    flow.nodes.find(
      (n) => n.type === "entry_command" || n.type === "entry_event"
    )?.data ?? {};

  return {
    name: bundle.name,
    settings: bundle.settings,
    items: [
      ...bundle.commands.map((c) => ({
        key: c.id,
        label: `/${String(entry(c.flow_source).name ?? "")}`,
      })),
      ...bundle.event_listeners.map((l) => {
        const type = String(entry(l.flow_source).event_type ?? "");
        return {
          key: l.id,
          label:
            (l.source === "discord" && discordEventType(type)?.label) ||
            capitalize(eventSourceName(l.source).singular),
        };
      }),
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
