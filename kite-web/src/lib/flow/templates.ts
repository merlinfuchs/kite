import {
  AlarmClockIcon,
  ArrowUp10Icon,
  BotOffIcon,
  BrainCircuitIcon,
  GavelIcon,
  GiftIcon,
  InfoIcon,
  LightbulbIcon,
  LucideIcon,
  MailIcon,
  MoonIcon,
  PackageIcon,
  PinIcon,
  ShieldAlertIcon,
  ShieldCheckIcon,
  SirenIcon,
  StarIcon,
  UserRoundPlusIcon,
} from "lucide-react";
import { z } from "zod";
import { FlowData } from "./dataSchema";
import { MessageData } from "../types/message.gen";
import { Bundle } from "../types/wire.gen";
import { parameterKeyPattern, referencedParameters } from "./parameters";
import userAvatar from "@/tools/common/utils/discordCdn";

export const templateIcons: Record<string, LucideIcon> = {
  gavel: GavelIcon,
  "brain-circuit": BrainCircuitIcon,
  "user-round-plus": UserRoundPlusIcon,
  mail: MailIcon,
  siren: SirenIcon,
  "shield-check": ShieldCheckIcon,
  "shield-alert": ShieldAlertIcon,
  lightbulb: LightbulbIcon,
  info: InfoIcon,
  "arrow-up-1-0": ArrowUp10Icon,
  star: StarIcon,
  moon: MoonIcon,
  gift: GiftIcon,
  pin: PinIcon,
  "alarm-clock": AlarmClockIcon,
  "bot-off": BotOffIcon,
};

export const templateCategories = [
  { value: "moderation", label: "Moderation" },
  { value: "ai", label: "AI" },
  { value: "community", label: "Community" },
  { value: "support", label: "Support" },
  { value: "games", label: "Games" },
  { value: "other", label: "Other" },
] as const;

export type TemplateCategory = (typeof templateCategories)[number]["value"];

const flowSourceSchema = z.object({
  nodes: z.array(
    z.object({
      id: z.string(),
      type: z.string(),
      position: z.object({ x: z.number(), y: z.number() }),
      data: z.record(z.string(), z.unknown()),
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

export const templateParameterSchema = z.object({
  key: z.string().regex(parameterKeyPattern),
  label: z.string().min(1),
  description: z.string().min(1),
  // channel and role are picked from a server the app is in, and their ID is
  // the value.
  type: z.enum(["text", "textarea", "channel", "role"]),
  required: z.boolean(),
  // The value it starts with when the template is added. Only text
  // parameters have one.
  default: z.string().optional(),
});

export type TemplateParameter = z.infer<typeof templateParameterSchema>;

// parameterDefaults are the values the parameters start with when the
// template is added, by their key.
export function parameterDefaults(
  parameters: TemplateParameter[]
): Record<string, string> {
  return Object.fromEntries(
    parameters.flatMap((p) => (p.default ? [[p.key, p.default]] : []))
  );
}

// Parameters of exported flows can be without a description, unlike the ones
// of the templates.
const bundleParameterSchema = templateParameterSchema.extend({
  description: z.string(),
});

// The commands, event listeners and messages have the shape of the editor's
// JSON export. Their blocks use {{params.<key>}}, and flows that use the same
// key get the same value when the template is added.
const itemIdSchema = z.string().regex(/^[a-z0-9-]+$/);

export const templateFileSchema = z.object({
  // Used in links and analytics, and as the bundle ID of installs, so it must
  // never change.
  id: z.string().regex(/^[a-z0-9-]+$/),
  // Goes up whenever the flows change, so apps that added the template can
  // update to it.
  version: z.number().int().min(1),
  // What changed in a version, which apps see when they update to it. A test
  // checks every version after the first has one.
  changelog: z
    .array(
      z.object({
        version: z.number().int().min(2),
        notes: z.string().min(1).max(1000),
      })
    )
    .optional(),
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
  parameters: z.array(templateParameterSchema),
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
  // is posted in the channel of that parameter when the template is added.
  // One with customize is written in the message editor when adding the
  // template, starting with data. Its buttons and select menus can't run
  // flows, the message editor there can't edit them.
  messages: z
    .array(
      z
        .object({
          id: itemIdSchema,
          name: z.string().min(1).max(100),
          description: z.string().max(255).nullable(),
          data: z.record(z.string(), z.unknown()),
          flow_sources: z.record(z.string(), flowSourceSchema),
          send_to: z.string().regex(parameterKeyPattern).optional(),
          customize: z.boolean().optional(),
        })
        .refine(
          (m) => !m.customize || Object.keys(m.flow_sources).length === 0,
          "A message with customize can't have flow_sources"
        )
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

// Handles the service writes from its own types, which can be null.
const handleSchema = z
  .string()
  .nullish()
  .transform((v) => v ?? undefined)
  .optional();

const bundleFlowSchema = z.object({
  parameters: z
    .array(
      bundleParameterSchema
        .partial({ label: true, description: true, required: true })
        .extend({ value: z.string() })
    )
    .optional(),
  nodes: z.array(
    z.object({
      id: z.string(),
      type: z.string(),
      position: z.object({ x: z.number(), y: z.number() }),
      data: z.record(z.string(), z.unknown()),
    })
  ),
  edges: z.array(
    z.object({
      id: z.string(),
      source: z.string(),
      target: z.string(),
      sourceHandle: handleSchema,
      targetHandle: handleSchema,
      type: handleSchema,
    })
  ),
});

// bundleContentSchema reads the content of a bundle from the service, an
// official template or an export of several items. It's the template file
// format without what's only for the site's pages.
export const bundleContentSchema = z.object({
  id: z.string(),
  version: z.number(),
  name: z.string(),
  description: z.string(),
  parameters: z.array(bundleParameterSchema),
  commands: z.array(
    z.object({ id: z.string(), flow_source: bundleFlowSchema })
  ),
  event_listeners: z.array(
    z.object({
      id: z.string(),
      source: z.string(),
      flow_source: bundleFlowSchema,
    })
  ),
  messages: z.array(
    z.object({
      id: z.string(),
      name: z.string(),
      description: z.string().nullable(),
      data: z.record(z.string(), z.unknown()),
      flow_sources: z.record(z.string(), bundleFlowSchema),
      send_to: z
        .string()
        .transform((v) => v || undefined)
        .optional(),
      customize: z.boolean().optional(),
    })
  ),
  variables: z.array(
    z.object({ id: z.string(), name: z.string(), scoped: z.boolean() })
  ),
});

export type BundleContent = z.infer<typeof bundleContentSchema>;

type BundleFlow = BundleContent["commands"][number]["flow_source"];

// A template file, or a bundle of the service with its listing.
type TemplateSource = Omit<BundleContent, "messages" | "variables"> & {
  messages?: BundleContent["messages"];
  variables?: BundleContent["variables"];
  icon: string;
  category?: TemplateCategory;
  about: string[];
  faq: TemplateFile["faq"];
  setup?: string[];
  // Only the files have it, for the credit calculator of the pricing page.
  usage?: TemplateFile["usage"];
  versionNotes?: Template["versionNotes"];
};

// TemplateAuthor is who made the template: Kite for official ones, a user for
// bundles.
export type TemplateAuthor =
  | { kind: "kite" }
  | { kind: "you" }
  | {
      kind: "user";
      discordId: string;
      name: string;
      username: string;
      avatarUrl: string;
    };

export type Template = {
  id: string;
  version: number;
  name: string;
  description: string;
  icon: LucideIcon;
  // Bundles of users have no category.
  category?: TemplateCategory;
  author: TemplateAuthor;
  // Whether it's in the marketplace, like official templates and the bundles
  // of users a review approved.
  listed: boolean;
  // What changed in the last versions, newest first.
  versionNotes: { version: number; notes: string }[];
  setup?: string[];
  about: string[];
  faq: TemplateFile["faq"];
  usage?: TemplateFile["usage"];
  parameters: TemplateParameter[];
  commands: {
    key: string;
    name: string;
    description: string;
    // The keys of the parameters it uses.
    parameterKeys: string[];
    flowSource: FlowData;
  }[];
  eventListeners: {
    key: string;
    source: string;
    type: string;
    description: string;
    parameterKeys: string[];
    flowSource: FlowData;
  }[];
  messages: {
    key: string;
    name: string;
    description: string | null;
    parameterKeys: string[];
    // The key of the channel parameter the message is posted in.
    sendTo?: string;
    // Whether it's written in the message editor when adding the template.
    customize: boolean;
    data: MessageData;
    flowSources: Record<string, FlowData>;
  }[];
  variables: { key: string; name: string; scoped: boolean }[];
};

// templatePath is the public page of a template or a listed bundle in the
// marketplace.
export function templatePath(id: string) {
  return `/marketplace/${id}`;
}

// creatorPath is the page of a user's templates in the marketplace. The
// Discord ID doesn't change like the username can.
export function creatorPath(discordId: string) {
  return `/marketplace/creators/${discordId}`;
}

// creatorBundles are the listed bundles of the user with the Discord ID.
export function creatorBundles(bundles: Bundle[], discordId: string) {
  return bundles.filter((b) => b.owner?.discord_id === discordId);
}

// templateFromBundle is the template of a bundle of the service, or undefined
// if it isn't one the site can show. Official ones always have a listing, the
// bundles of users once they're listed in the marketplace.
export function templateFromBundle(bundle: Bundle): Template | undefined {
  const content = bundleContentSchema.safeParse(bundle.content);
  if (!content.success) return undefined;
  const versionNotes = bundle.version_notes ?? [];
  const category = templateCategories.find(
    (c) => c.value === bundle.listing?.category
  );
  const listing =
    bundle.listing && category
      ? { ...bundle.listing, category: category.value }
      : undefined;

  if (bundle.official) {
    if (!listing) return undefined;
    return loadTemplate(
      { ...content.data, ...listing, versionNotes },
      { kind: "kite" },
      true
    );
  }
  return loadTemplate(
    {
      ...content.data,
      icon: "package",
      about: [],
      faq: [],
      ...listing,
      versionNotes,
    },
    bundle.yours || !bundle.owner
      ? { kind: "you" }
      : {
          kind: "user",
          discordId: bundle.owner.discord_id,
          name: bundle.owner.display_name,
          username: bundle.owner.discord_username,
          avatarUrl: userAvatar({
            id: bundle.owner.discord_id,
            avatar: bundle.owner.discord_avatar,
          }),
        },
    !!listing
  );
}

export function loadTemplate(
  file: TemplateSource,
  author: TemplateAuthor,
  listed: boolean
): Template {
  return {
    listed,
    id: file.id,
    version: file.version,
    name: file.name,
    description: file.description,
    icon: templateIcons[file.icon] ?? PackageIcon,
    category: file.category,
    author,
    versionNotes: file.versionNotes ?? [],
    setup: file.setup,
    about: file.about,
    faq: file.faq,
    usage: file.usage,
    parameters: file.parameters,
    commands: file.commands.map(({ id, flow_source }) => {
      const entry = entryData(flow_source);
      return {
        key: id,
        name: String(entry.name ?? ""),
        description: String(entry.description ?? ""),
        parameterKeys: referencedParameters(flow_source),
        flowSource: flow_source as FlowData,
      };
    }),
    eventListeners: file.event_listeners.map(({ id, source, flow_source }) => {
      const entry = entryData(flow_source);
      return {
        key: id,
        source,
        type: String(entry.event_type ?? ""),
        description: String(entry.description ?? ""),
        parameterKeys: referencedParameters(flow_source),
        flowSource: flow_source as FlowData,
      };
    }),
    messages: (file.messages ?? []).map((message) => {
      const flows = Object.values(message.flow_sources);
      return {
        key: message.id,
        name: message.name,
        description: message.description,
        parameterKeys: Array.from(
          new Set([
            ...flows.flatMap((flow) => referencedParameters(flow)),
            ...(message.send_to ? [message.send_to] : []),
          ])
        ),
        sendTo: message.send_to,
        customize: message.customize ?? false,
        data: message.data as MessageData,
        flowSources: message.flow_sources as Record<string, FlowData>,
      };
    }),
    variables: (file.variables ?? []).map(({ id, name, scoped }) => ({
      key: id,
      name,
      scoped,
    })),
  };
}

// entryData is the data of the block that starts the flow.
export function entryData(flow: BundleFlow) {
  return (
    flow.nodes.find(
      (n) => n.type === "entry_command" || n.type === "entry_event"
    )?.data ?? {}
  );
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

// templateFlows lists every flow of a template with what runs it.
export function templateFlows(template: Template): TemplateFlow[] {
  return [
    ...template.commands.map((c) => ({
      key: c.key,
      kind: "command" as const,
      name: c.name,
      flow: c.flowSource,
    })),
    ...template.eventListeners.map((l) => ({
      key: l.key,
      kind: "event" as const,
      eventType: l.type,
      flow: l.flowSource,
    })),
    ...template.messages.flatMap((message) =>
      Object.entries(message.flowSources).map(([id, flow]) => ({
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
