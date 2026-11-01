import {
  AlarmClockIcon,
  ArrowUp10Icon,
  BotOffIcon,
  BrainCircuitIcon,
  GavelIcon,
  GiftIcon,
  InfoIcon,
  LanguagesIcon,
  LightbulbIcon,
  LucideIcon,
  MailIcon,
  MoonIcon,
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

// templateIcons are the icons of Kite's templates, which are bundled. User
// templates can have any Lucide icon, see TemplateIcon.
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
  languages: LanguagesIcon,
};

// lucideName is the name of the Lucide component of the icon, like
// ArrowUp10 for arrow-up-1-0.
export function lucideName(icon: string) {
  return icon
    .split("-")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join("");
}

// iconName is the name a template stores for the Lucide component, which
// lucideName turns back into it.
export function iconName(component: string) {
  return component.replace(/(?<!^)([A-Z])/g, "-$1").toLowerCase();
}

// Each category has a page in the marketplace, with the title and
// description. The colors are of the icons of the templates in a category, so
// the cards of the marketplace don't all look the same.
export const templateCategories = [
  {
    value: "moderation",
    label: "Moderation",
    title: "Discord Moderation Bot Templates",
    description:
      "Commands to ban, kick and mute members, and automods that remove spam and rule breaks for you.",
    icon: "text-blue-600 dark:text-blue-400",
    tile: "bg-blue-500/10",
    banner: "from-blue-500/25 via-blue-500/5",
  },
  {
    value: "ai",
    label: "AI",
    title: "Discord AI Bot Templates",
    description:
      "Commands that answer the questions of your members with AI, or use it to do things for them.",
    icon: "text-violet-600 dark:text-violet-400",
    tile: "bg-violet-500/10",
    banner: "from-violet-500/25 via-violet-500/5",
  },
  {
    value: "welcome",
    label: "Welcome",
    title: "Discord Welcome and Verification Bot Templates",
    description:
      "Greet new members, have them verify or register, and reward the ones who invite others.",
    icon: "text-cyan-600 dark:text-cyan-400",
    tile: "bg-cyan-500/10",
    banner: "from-cyan-500/25 via-cyan-500/5",
  },
  {
    value: "community",
    label: "Community",
    title: "Discord Community Bot Templates",
    description:
      "Suggestions, giveaways, starboards and more to keep your server active.",
    icon: "text-emerald-600 dark:text-emerald-400",
    tile: "bg-emerald-500/10",
    banner: "from-emerald-500/25 via-emerald-500/5",
  },
  {
    value: "utility",
    label: "Utility",
    title: "Discord Utility Bot Templates",
    description:
      "Server and user info, reminders, notes, translations and other everyday commands.",
    icon: "text-slate-600 dark:text-slate-300",
    tile: "bg-slate-500/10",
    banner: "from-slate-500/25 via-slate-500/5",
  },
  {
    value: "support",
    label: "Support",
    title: "Discord Ticket and Modmail Bot Templates",
    description:
      "Tickets and modmail, so members can reach your staff in private.",
    icon: "text-amber-600 dark:text-amber-400",
    tile: "bg-amber-500/10",
    banner: "from-amber-500/25 via-amber-500/5",
  },
  {
    value: "games",
    label: "Fun & Games",
    title: "Discord Fun and Game Bot Templates",
    description:
      "Games, fun commands and random activities for your members, and panels for ER:LC servers.",
    icon: "text-pink-600 dark:text-pink-400",
    tile: "bg-pink-500/10",
    banner: "from-pink-500/25 via-pink-500/5",
  },
  {
    value: "other",
    label: "Other",
    title: "More Discord Bot Templates",
    description:
      "Commands that fit no other category, like leave requests for staff.",
    icon: "text-primary",
    tile: "bg-primary/10",
    banner: "from-primary/25 via-primary/5",
  },
] as const;

// templateCategory is the category of a template, Other if it has none.
export function templateCategory(template: Template) {
  return (
    templateCategories.find((c) => c.value === template.category) ??
    templateCategories[templateCategories.length - 1]!
  );
}

// featuredTemplates are the templates the marketplace shows first, two of
// Kite's and one a user made, picked again every day. day is the number of
// days since 1970 in UTC, so everyone sees the same ones.
export function featuredTemplates(templates: Template[], day: number) {
  // A different order of them every day.
  const pick = (pool: Template[], count: number) =>
    pool
      .map((t) => ({ t, rank: hash(`${day}:${t.id}`) }))
      .sort((a, b) => a.rank - b.rank)
      .slice(0, count)
      .map(({ t }) => t);

  const community = pick(
    templates.filter(
      (t) =>
        t.author.kind === "user" &&
        t.listed &&
        (t.installs ?? 0) >= minFeaturedInstalls
    ),
    1
  );
  const official = pick(
    templates.filter((t) => t.author.kind === "kite"),
    3 - community.length
  );
  return [...official.slice(0, 1), ...community, ...official.slice(1)];
}

// Users' templates need a few installs to be featured, so the ones nobody
// tried yet don't get the top of the marketplace.
const minFeaturedInstalls = 5;

// utcDay is the day featuredTemplates picks for.
export function utcDay(date = new Date()) {
  return Math.floor(date.getTime() / 86_400_000);
}

// hash is FNV-1a, enough to shuffle the featured templates.
function hash(text: string) {
  let h = 2166136261;
  for (let i = 0; i < text.length; i++) {
    h ^= text.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

export type TemplateCategory = (typeof templateCategories)[number]["value"];

const flowNodeSchema = z.object({
  id: z.string(),
  type: z.string(),
  position: z.object({ x: z.number(), y: z.number() }),
  // The size of notes, they're a dot without it.
  width: z.number().optional(),
  height: z.number().optional(),
  data: z.record(z.string(), z.unknown()),
});

const flowSourceSchema = z.object({
  nodes: z.array(flowNodeSchema),
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
  // the value. boolean is yes or no, picked from a menu, as true or false.
  type: z.enum(["text", "textarea", "channel", "role", "boolean"]),
  required: z.boolean(),
  // The value it starts with when the template is added. Channels and roles
  // don't have one.
  default: z.string().optional(),
});

export type TemplateParameter = z.infer<typeof templateParameterSchema>;

export const parameterTypeLabels: Record<TemplateParameter["type"], string> = {
  text: "Text",
  textarea: "Long text",
  channel: "Channel",
  role: "Role",
  boolean: "Yes/No",
};

const isBooleanValue = (value: string) => value === "true" || value === "false";

// formatParameterValue is the value as it's shown, Yes or No for a yes/no.
export function formatParameterValue(type: string, value: string): string {
  if (type !== "boolean" || !isBooleanValue(value)) return value;
  return value === "true" ? "Yes" : "No";
}

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
  parameters: z.array(
    templateParameterSchema.refine(
      (p) => p.type !== "boolean" || !p.default || isBooleanValue(p.default),
      { message: "The default of a yes/no must be true or false" }
    )
  ),
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
  nodes: z.array(flowNodeSchema),
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
  aiUse?: "assisted" | "generated";
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
  // icon is the name of a Lucide icon, in kebab case.
  icon: string;
  // Bundles of users have no category.
  category?: TemplateCategory;
  author: TemplateAuthor;
  // Whether it's in the marketplace, like official templates and the bundles
  // of users a review approved.
  listed: boolean;
  // AI contribution to a user's template, measured from AI-tagged blocks.
  aiUse?: "assisted" | "generated";
  // What changed in the last versions, newest first.
  versionNotes: { version: number; notes: string }[];
  // How many apps added it, unknown for the official ones from the files.
  installs?: number;
  // How many users starred it, unknown for the official ones from the files.
  stars?: number;
  // When it was made, unknown for the official ones from the files.
  createdAt?: string;
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

// relatedTemplates are the templates the page of one suggests: more by the
// user who made it, and the most added others of its category.
export function relatedTemplates(template: Template, templates: Template[]) {
  const others = templates.filter((t) => t.id !== template.id && t.listed);
  const creator =
    template.author.kind === "user" ? template.author.discordId : undefined;
  const byCreator = creator
    ? others
        .filter(
          (t) => t.author.kind === "user" && t.author.discordId === creator
        )
        .slice(0, 3)
    : [];
  const similar = others
    .filter((t) => t.category === template.category && !byCreator.includes(t))
    .sort((a, b) => (b.installs ?? 0) - (a.installs ?? 0))
    .slice(0, 3);
  return { byCreator, similar };
}

// templatePath is the public page of a template or a listed bundle in the
// marketplace.
export function templatePath(id: string) {
  return `/marketplace/${id}`;
}

// categoryPath is the page of the templates of a category in the
// marketplace.
export function categoryPath(category: TemplateCategory) {
  return `/marketplace/categories/${category}`;
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

// Below this, a star count makes a template look unused, so it isn't shown.
export const minShownStars = 5;

// templateFromBundle is the template of a bundle of the service, or undefined
// if it isn't one the site can show. Official ones always have a listing,
// users' ones show their page until a listing of them is approved.
export function templateFromBundle(bundle: Bundle): Template | undefined {
  const content = bundleContentSchema.safeParse(bundle.content);
  if (!content.success) return undefined;
  const listing = bundle.listing;
  const category = templateCategories.find(
    (c) => c.value === listing?.category
  )?.value;
  // The content has the name and description it was published with, which
  // can have changed since.
  const source = {
    ...content.data,
    name: bundle.name,
    description: bundle.description,
    icon: listing?.icon || "package",
    category,
    about: listing?.about ?? [],
    faq: listing?.faq ?? [],
    setup: listing?.setup,
    versionNotes: bundle.version_notes ?? [],
    aiUse: marketplaceAIUse(bundle.ai_use),
  };

  if (bundle.official) {
    if (!listing || !category) return undefined;
    return {
      ...loadTemplate(source, { kind: "kite" }, true),
      installs: bundle.install_count,
      stars: bundle.star_count,
      createdAt: bundle.created_at,
    };
  }
  return {
    ...loadTemplate(
      source,
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
      bundle.listed
    ),
    installs: bundle.install_count,
    stars: bundle.star_count,
    createdAt: bundle.created_at,
  };
}

function marketplaceAIUse(value: string | undefined): Template["aiUse"] {
  if (value === "assisted" || value === "generated") {
    return value;
  }
}

export function loadTemplate(
  file: TemplateSource,
  author: TemplateAuthor,
  listed: boolean
): Template {
  return {
    listed,
    aiUse: file.aiUse,
    id: file.id,
    version: file.version,
    name: file.name,
    description: file.description,
    icon: file.icon,
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
  const component = findFlowSourceComponent(data, flowSourceId);
  const label =
    component && String(component.label ?? component.placeholder ?? "");
  return label ? `"${label}"` : undefined;
}

// findFlowSourceComponent finds the button or select menu in a message that
// runs a flow. Components can be nested in rows, sections and containers.
export function findFlowSourceComponent(
  value: unknown,
  flowSourceId: string
): Record<string, unknown> | undefined {
  if (!value || typeof value !== "object") return undefined;
  const component = value as Record<string, unknown>;
  if (component.flow_source_id === flowSourceId) return component;
  for (const child of Object.values(component)) {
    const found = findFlowSourceComponent(child, flowSourceId);
    if (found) return found;
  }
  return undefined;
}
