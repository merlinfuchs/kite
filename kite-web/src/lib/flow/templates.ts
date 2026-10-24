import {
  BrainCircuitIcon,
  GavelIcon,
  InfoIcon,
  LightbulbIcon,
  LucideIcon,
  MailIcon,
  ShieldCheckIcon,
  SirenIcon,
  UserRoundPlusIcon,
} from "lucide-react";
import { z } from "zod";
import { FlowData } from "./dataSchema";
import { MessageData } from "../types/message.gen";
import askAi from "../templates/ask-ai.json";
import erlc from "../templates/erlc.json";
import moderation from "../templates/moderation.json";
import modmail from "../templates/modmail.json";
import serverInfo from "../templates/server-info.json";
import suggestions from "../templates/suggestions.json";
import verification from "../templates/verification.json";
import welcomer from "../templates/welcomer.json";

// The templates in the order of the Templates page. Each file is described by
// templateFileSchema, see "Adding a template" in AGENTS.md.
export const templateFiles: unknown[] = [
  moderation,
  askAi,
  welcomer,
  verification,
  suggestions,
  serverInfo,
  modmail,
  erlc,
];

// $setting.<key> in a template, see templateFileSchema.
const settingPattern = /\$setting\.([a-z0-9_]+)/g;
const onlySettingPattern = /^\$setting\.([a-z0-9_]+)$/;

const templateIcons: Record<string, LucideIcon> = {
  gavel: GavelIcon,
  "brain-circuit": BrainCircuitIcon,
  "user-round-plus": UserRoundPlusIcon,
  mail: MailIcon,
  siren: SirenIcon,
  "shield-check": ShieldCheckIcon,
  lightbulb: LightbulbIcon,
  info: InfoIcon,
};

export const templateCategories = [
  { value: "moderation", label: "Moderation" },
  { value: "ai", label: "AI" },
  { value: "community", label: "Community" },
  { value: "support", label: "Support" },
  { value: "games", label: "Games" },
] as const;

export type TemplateCategory = (typeof templateCategories)[number]["value"];

const flowSourceSchema = z.object({
  nodes: z.array(
    z.object({
      id: z.string(),
      type: z.string(),
      position: z.object({ x: z.number(), y: z.number() }),
      data: z.record(z.unknown()),
    })
  ),
  edges: z.array(
    z.object({
      id: z.string(),
      source: z.string(),
      target: z.string(),
      sourceHandle: z.string().optional(),
      targetHandle: z.string().optional(),
      type: z.string().optional(),
    })
  ),
});

const templateSettingSchema = z.object({
  key: z.string().regex(/^[a-z0-9_]+$/),
  label: z.string().min(1),
  description: z.string().min(1),
  // channel and role are picked from a server the app is in, and their ID is
  // the value.
  type: z.enum(["text", "textarea", "channel", "role"]),
  required: z.boolean(),
});

export type TemplateSetting = z.infer<typeof templateSettingSchema>;

// The commands, event listeners and messages have the shape of the editor's
// JSON export. $setting.<key> in their blocks and messages is replaced with
// the value of the setting when the template is added.
const itemIdSchema = z.string().regex(/^[a-z0-9-]+$/);

export const templateFileSchema = z.object({
  // Used in links and analytics, and as the bundle ID of installs, so it must
  // never change.
  id: z.string().regex(/^[a-z0-9-]+$/),
  // Goes up whenever the flows change, so apps that added the template can
  // update to it.
  version: z.number().int().min(1),
  name: z.string().min(1),
  description: z.string().min(1),
  // Paragraphs on what the template does and what it needs, and questions
  // people search for. They're on the template's public page.
  about: z.array(z.string().min(1)).min(1),
  faq: z
    .array(z.object({ question: z.string().min(1), answer: z.string().min(1) }))
    .min(1),
  // For the credit calculator on the pricing page: what one run of the
  // template is, and a guess of how many there are a month per 1,000 members.
  usage: z.object({
    unit: z.string().min(1),
    unit_one: z.string().min(1),
    per_thousand_members: z.number().min(0),
  }),
  icon: z.string().refine((icon) => icon in templateIcons, "unknown icon"),
  category: z.enum(
    templateCategories.map((c) => c.value) as [
      TemplateCategory,
      ...TemplateCategory[]
    ]
  ),
  // The steps outside of the template that it needs to work.
  setup: z.array(z.string()).optional(),
  settings: z.array(templateSettingSchema),
  // Item IDs match what an app installed to a later version of the
  // template, so they must never change either.
  commands: z.array(
    z.object({ id: itemIdSchema, flow_source: flowSourceSchema })
  ),
  event_listeners: z.array(
    z.object({
      id: itemIdSchema,
      source: z.literal("discord"),
      flow_source: flowSourceSchema,
    })
  ),
  // Message templates, which blocks use by their id. A message with send_to
  // is posted in the channel of that setting when the template is added.
  messages: z
    .array(
      z.object({
        id: itemIdSchema,
        name: z.string().min(1).max(100),
        description: z.string().max(255).nullable(),
        data: z.record(z.unknown()),
        flow_sources: z.record(flowSourceSchema),
        send_to: z.string().regex(onlySettingPattern).optional(),
      })
    )
    .optional(),
  // Stored variables, which blocks use by their id. They're created when the
  // template is added, or when an update uses one the app doesn't have.
  variables: z
    .array(
      z.object({
        id: itemIdSchema,
        name: z
          .string()
          .min(1)
          .max(100)
          .regex(/^[a-zA-Z0-9_]+$/),
        scoped: z.boolean(),
      })
    )
    .optional(),
});

export type TemplateFile = z.infer<typeof templateFileSchema>;

export type Template = {
  id: string;
  version: number;
  name: string;
  description: string;
  icon: LucideIcon;
  category: TemplateCategory;
  setup?: string[];
  about: string[];
  faq: TemplateFile["faq"];
  usage: TemplateFile["usage"];
  settings: TemplateSetting[];
  commands: {
    key: string;
    name: string;
    description: string;
    // The keys of the settings it uses.
    settingKeys: string[];
    flowSource(settings: Record<string, string>): FlowData;
  }[];
  eventListeners: {
    key: string;
    source: string;
    type: string;
    description: string;
    settingKeys: string[];
    flowSource(settings: Record<string, string>): FlowData;
  }[];
  messages: {
    key: string;
    name: string;
    description: string | null;
    settingKeys: string[];
    // The key of the channel setting the message is posted in.
    sendTo?: string;
    data(settings: Record<string, string>): MessageData;
    flowSources(settings: Record<string, string>): Record<string, FlowData>;
  }[];
  variables: { key: string; name: string; scoped: boolean }[];
};

const templates = (templateFiles as TemplateFile[]).map(loadTemplate);

export function getTemplates(): Template[] {
  return templates;
}

// templatePaths are the public pages of the templates.
export function templatePaths() {
  return templates.map((t) => `/templates/${t.id}`);
}

export function getTemplate(id: string) {
  return templates.find((t) => t.id === id);
}

function loadTemplate(file: TemplateFile): Template {
  return {
    id: file.id,
    version: file.version,
    name: file.name,
    description: file.description,
    icon: templateIcons[file.icon],
    category: file.category,
    setup: file.setup,
    about: file.about,
    faq: file.faq,
    usage: file.usage,
    settings: file.settings,
    commands: file.commands.map(({ id, flow_source }) => {
      const entry = entryData(flow_source);
      return {
        key: id,
        name: String(entry.name ?? ""),
        description: String(entry.description ?? ""),
        settingKeys: usedSettingKeys(flow_source),
        flowSource: (settings) => applyTemplateSettings(flow_source, settings),
      };
    }),
    eventListeners: file.event_listeners.map(({ id, source, flow_source }) => {
      const entry = entryData(flow_source);
      return {
        key: id,
        source,
        type: String(entry.event_type ?? ""),
        description: String(entry.description ?? ""),
        settingKeys: usedSettingKeys(flow_source),
        flowSource: (settings) => applyTemplateSettings(flow_source, settings),
      };
    }),
    messages: (file.messages ?? []).map((message) => ({
      key: message.id,
      name: message.name,
      description: message.description,
      settingKeys: usedSettingKeys(message),
      sendTo: message.send_to?.match(onlySettingPattern)?.[1],
      data: (settings) => applySettings(message.data, settings) as MessageData,
      flowSources: (settings) =>
        Object.fromEntries(
          Object.entries(message.flow_sources).map(([id, flow]) => [
            id,
            applyTemplateSettings(flow, settings),
          ])
        ),
    })),
    variables: (file.variables ?? []).map(({ id, name, scoped }) => ({
      key: id,
      name,
      scoped,
    })),
  };
}

function entryData(flow: TemplateFile["commands"][number]["flow_source"]) {
  return (
    flow.nodes.find(
      (n) => n.type === "entry_command" || n.type === "entry_event"
    )?.data ?? {}
  );
}

// templateSettingKeys are the settings a value of a block uses.
export function templateSettingKeys(value: string) {
  return Array.from(value.matchAll(settingPattern), (m) => m[1]);
}

function usedSettingKeys(item: unknown) {
  return Array.from(new Set(templateSettingKeys(JSON.stringify(item))));
}

function applyTemplateSettings(
  flow: TemplateFile["commands"][number]["flow_source"],
  settings: Record<string, string>
): FlowData {
  return {
    nodes: flow.nodes.map((node) => ({
      ...node,
      data: applySettings(
        node.data,
        settings
      ) as FlowData["nodes"][number]["data"],
    })),
    edges: flow.edges,
  };
}

function applySettings(
  value: unknown,
  settings: Record<string, string>
): unknown {
  if (typeof value === "string") {
    return value.replace(settingPattern, (_, key) => settings[key] ?? "");
  }
  if (Array.isArray(value)) {
    return value.map((v) => applySettings(v, settings));
  }
  if (value && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value).flatMap(([key, v]) => {
        // A value that is only an optional setting that wasn't filled in is
        // left out, so the block uses its default.
        const only = typeof v === "string" && v.match(onlySettingPattern);
        if (only && !settings[only[1]]) return [];
        return [[key, applySettings(v, settings)]];
      })
    );
  }
  return value;
}

export type TemplateFlow = { key: string; flow: FlowData } & (
  | { kind: "command"; name: string }
  | { kind: "event"; eventType: string }
  | {
      kind: "button";
      message: Template["messages"][number];
      flowSourceId: string;
    }
);

// templateFlows lists every flow of a template with what runs it, without the
// values of the settings.
export function templateFlows(template: Template): TemplateFlow[] {
  return [
    ...template.commands.map((c) => ({
      key: c.key,
      kind: "command" as const,
      name: c.name,
      flow: c.flowSource({}),
    })),
    ...template.eventListeners.map((l) => ({
      key: l.key,
      kind: "event" as const,
      eventType: l.type,
      flow: l.flowSource({}),
    })),
    ...template.messages.flatMap((message) =>
      Object.entries(message.flowSources({})).map(([id, flow]) => ({
        key: `${message.key}:${id}`,
        kind: "button" as const,
        message,
        flowSourceId: id,
        flow,
      }))
    ),
  ];
}

// componentLabel is the label of the button or select menu in a message that
// runs a flow.
export function componentLabel(data: unknown, flowSourceId: string) {
  let label: string | undefined;
  const walk = (value: unknown) => {
    if (label || !value || typeof value !== "object") return;
    const component = value as Record<string, unknown>;
    if (component.flow_source_id === flowSourceId) {
      label = String(component.label ?? component.placeholder ?? "");
      return;
    }
    Object.values(component).forEach(walk);
  };
  walk(data);
  return label ? `"${label}"` : undefined;
}
