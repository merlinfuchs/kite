import {
  BrainCircuitIcon,
  GavelIcon,
  LucideIcon,
  MailIcon,
  SirenIcon,
  UserRoundPlusIcon,
} from "lucide-react";
import { z } from "zod";
import { FlowData } from "./dataSchema";
import askAi from "../templates/ask-ai.json";
import erlc from "../templates/erlc.json";
import moderation from "../templates/moderation.json";
import modmail from "../templates/modmail.json";
import welcomer from "../templates/welcomer.json";

// The templates in the order of the Templates page. Each file is described by
// templateFileSchema, see "Adding a template" in AGENTS.md.
export const templateFiles: unknown[] = [
  moderation,
  askAi,
  welcomer,
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

// The commands and event listeners have the shape of the editor's JSON
// export. $setting.<key> in their blocks is replaced with the value of the
// setting when the template is added.
const itemIdSchema = z.string().regex(/^[a-z0-9-]+$/);

export const templateFileSchema = z.object({
  // Used in links and analytics, and as the bundle ID of installs, so it must
  // never change.
  id: z.string().regex(/^[a-z0-9-]+$/),
  // Goes up whenever the flows change. Installed copies don't update.
  version: z.number().int().min(1),
  name: z.string().min(1),
  description: z.string().min(1),
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
  settings: TemplateSetting[];
  // settingKeys of an item are the settings its blocks use.
  commands: {
    key: string;
    name: string;
    description: string;
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
};

const templates = (templateFiles as TemplateFile[]).map(loadTemplate);

export function getTemplates(): Template[] {
  return templates;
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
    settings: file.settings,
    commands: file.commands.map(({ id, flow_source }) => {
      const entry = entryData(flow_source);
      return {
        key: id,
        name: String(entry.name ?? ""),
        description: String(entry.description ?? ""),
        settingKeys: templateSettingKeys(JSON.stringify(flow_source)),
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
        settingKeys: templateSettingKeys(JSON.stringify(flow_source)),
        flowSource: (settings) => applyTemplateSettings(flow_source, settings),
      };
    }),
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

function applyTemplateSettings(
  flow: TemplateFile["commands"][number]["flow_source"],
  settings: Record<string, string>
): FlowData {
  const apply = (value: unknown): unknown => {
    if (typeof value === "string") {
      return value.replace(settingPattern, (_, key) => settings[key] ?? "");
    }
    if (Array.isArray(value)) {
      return value.map(apply);
    }
    if (value && typeof value === "object") {
      return Object.fromEntries(
        Object.entries(value).flatMap(([key, v]) => {
          // A value that is only an optional setting that wasn't filled in is
          // left out, so the block uses its default.
          const only = typeof v === "string" && v.match(onlySettingPattern);
          if (only && !settings[only[1]]) return [];
          return [[key, apply(v)]];
        })
      );
    }
    return value;
  };

  return {
    nodes: flow.nodes.map((node) => ({
      ...node,
      data: apply(node.data) as FlowData["nodes"][number]["data"],
    })),
    edges: flow.edges,
  };
}
