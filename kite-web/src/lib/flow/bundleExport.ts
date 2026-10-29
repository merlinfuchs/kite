import { LucideIcon, SlashSquareIcon } from "lucide-react";
import { FlowData } from "../types/flow.gen";
import { FlowContextType } from "./context";
import { eventListenerLabel, eventSourceName } from "./eventTypes";
import { bundleContentSchema, entryData, TemplateParameter } from "./templates";

export type ExportedBundle = {
  name: string;
  parameters: TemplateParameter[];
  // Commands and event listeners can be left out of an import.
  items: {
    key: string;
    name: string;
    description: string;
    icon: LucideIcon;
    context: FlowContextType;
    flow: FlowData;
  }[];
  // Message templates and variables are always imported, the blocks use them.
  messages: {
    key: string;
    name: string;
    description: string | null;
    data: Record<string, unknown>;
    flowSources: Record<string, FlowData>;
  }[];
  variables: { key: string; name: string; scoped: boolean }[];
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
    parameters: bundle.parameters,
    items: [
      ...bundle.commands.map((c) => ({
        key: c.id,
        name: `/${String(entryData(c.flow_source).name ?? "")}`,
        description: String(entryData(c.flow_source).description ?? ""),
        icon: SlashSquareIcon,
        context: "command" as const,
        flow: c.flow_source as FlowData,
      })),
      ...bundle.event_listeners.map((l) => ({
        key: l.id,
        name: eventListenerLabel(
          l.source,
          String(entryData(l.flow_source).event_type ?? "")
        ),
        description: String(entryData(l.flow_source).description ?? ""),
        icon: eventSourceName(l.source).icon,
        context: eventFlowContext(l.source),
        flow: l.flow_source as FlowData,
      })),
    ],
    messages: bundle.messages.map((m) => ({
      key: m.id,
      name: m.name,
      description: m.description,
      data: m.data,
      flowSources: m.flow_sources as Record<string, FlowData>,
    })),
    variables: bundle.variables.map((v) => ({
      key: v.id,
      name: v.name,
      scoped: v.scoped,
    })),
    flows: [
      ...bundle.commands.map((c) => c.flow_source),
      ...bundle.event_listeners.map((l) => l.flow_source),
      ...bundle.messages.flatMap((m) => Object.values(m.flow_sources)),
    ] as FlowData[],
  };
}

// eventFlowContext is the context of the flow of an event listener of a
// source, which labels its entry block.
export function eventFlowContext(source: string): FlowContextType {
  if (source === "schedule") return "event_schedule";
  if (source === "webhook") return "event_webhook";
  return "event_discord";
}
