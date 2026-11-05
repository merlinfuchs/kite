import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import {
  MessageMockup,
  VariableMockup,
} from "@/components/home/HomeFeaturesSection";
import {
  ProfileMockup,
  rotatingStatuses,
} from "@/components/home/ProfileMockups";
import { openSourcePoint } from "@/components/landing/copy";
import { Button } from "@/components/ui/button";
import { blockDefinitions } from "@/lib/blocks";
import { BlockDefinition } from "@/lib/blocks/types";
import env from "@/lib/env/client";
import { discordEventTypes } from "@/lib/flow/eventTypes";
import { faqJsonLd } from "@/lib/seo";
import { cn } from "@/lib/utils";
import { BillingPlan } from "@/lib/types/wire.gen";
import {
  fetchBundles,
  fetchPlans,
  fetchStats,
  withApiData,
} from "@/lib/apiData";
import {
  ArrowRightIcon,
  BotMessageSquareIcon,
  BrainIcon,
  CalculatorIcon,
  CircleDotIcon,
  ClockIcon,
  CodeIcon,
  CookieIcon,
  DatabaseIcon,
  Gamepad2Icon,
  GiftIcon,
  GlobeIcon,
  HandHeartIcon,
  LayoutGridIcon,
  LayoutTemplateIcon,
  LucideIcon,
  MousePointerClickIcon,
  PlugIcon,
  Repeat2Icon,
  SendIcon,
  ServerIcon,
  ShieldIcon,
  SirenIcon,
  SparklesIcon,
  SplitIcon,
  SquareSlashIcon,
  TicketIcon,
  UserRoundIcon,
  UsersIcon,
  WebhookIcon,
  ZapIcon,
} from "lucide-react";
import { GetStaticProps } from "next";
import Link from "next/link";
import { ReactNode } from "react";

interface Block {
  title: string;
  description: string;
  href: string;
  // The cheapest plan that has the block, or null for every plan.
  plan: string | null;
}

interface Point {
  title: string;
  description: string;
  href?: string;
  icon?: LucideIcon;
}

// What most people make, each linked to its template or landing page.
const starters: Point[] = [
  {
    title: "Moderation",
    description:
      "Ban, kick, and mute commands only your staff can use, plus AI that deletes spam.",
    href: "/discord-moderation-bot",
    icon: ShieldIcon,
  },
  {
    title: "AI chatbot",
    description:
      "Members ask your bot questions with /ask or by mentioning it.",
    href: "/ai-discord-bot",
    icon: BotMessageSquareIcon,
  },
  {
    title: "Welcome messages",
    description:
      "Greet new members in a channel you choose, and say goodbye when they leave.",
    href: "/discord-welcome-bot",
    icon: HandHeartIcon,
  },
  {
    title: "Giveaways",
    description:
      "A /giveaway command with an Enter button that picks random winners.",
    href: "/marketplace/giveaway",
    icon: GiftIcon,
  },
  {
    title: "Tickets and modmail",
    description: "Members open a ticket or DM your bot to reach your staff.",
    href: "/marketplace/categories/support",
    icon: TicketIcon,
  },
  {
    title: "ER:LC server",
    description:
      "Show your private server's players and run commands on it from Discord.",
    href: "/erlc-discord-bot",
    icon: SirenIcon,
  },
];

interface Props {
  blockCount: number;
  eventCount: number;
  categories: { title: string; blocks: Block[] }[];
  paidBlocks: string[];
  templates: {
    count: number;
    official: number;
    top: {
      id: string;
      name: string;
      description: string;
      installs: number;
    }[];
  } | null;
  bots: number | null;
  // The limits of the free plan, or null when the page is built without the
  // plans.
  free: {
    servers: number;
    credits: number;
    collaborators: number;
    aiPrompts: number;
    scheduleMinutes: number;
    // The cheapest plan that runs scheduled tasks more often.
    fasterSchedule: { plan: string; minutes: number } | null;
  } | null;
}

// The blocks come from their definitions, so the page lists every block the
// editor has without adding them to the page's bundle.
export const getStaticProps: GetStaticProps<Props> = async () => {
  const [bundles, plans, stats] = await Promise.all([
    fetchBundles(),
    fetchPlans(),
    fetchStats(),
  ]);

  const paidPlans = (plans ?? [])
    .flatMap((p) => (p && !p.hidden ? [p] : []))
    .sort((a, b) => a.price - b.price);
  const planFor = (block: BlockDefinition) => {
    if (!block.premium_feature) return null;
    const key = `feature_${block.premium_feature}` as keyof BillingPlan;
    const plan = paidPlans.find((p) => p[key] === true);
    if (!plan) return null;
    return plan === paidPlans[paidPlans.length - 1]
      ? plan.title
      : `${plan.title} and up`;
  };

  const freePlan = (plans ?? []).find((p) => p?.default);
  const fasterSchedule =
    freePlan &&
    paidPlans.find(
      (p) =>
        p.feature_min_schedule_interval_seconds <
        freePlan.feature_min_schedule_interval_seconds
    );
  const free: Props["free"] = freePlan
    ? {
        servers: freePlan.feature_max_guilds,
        credits: freePlan.feature_usage_credits_per_month,
        collaborators: freePlan.feature_max_collaborators,
        aiPrompts: freePlan.feature_max_ai_prompts_per_month,
        scheduleMinutes: freePlan.feature_min_schedule_interval_seconds / 60,
        fasterSchedule: fasterSchedule
          ? {
              plan: fasterSchedule.title,
              minutes:
                fasterSchedule.feature_min_schedule_interval_seconds / 60,
            }
          : null,
      }
    : null;

  // Blocks without a category are only created with other blocks, like the
  // branches of a condition, and aren't in the block list of the editor.
  const listed = blockDefinitions.filter((b) => b.category);
  const categories: Props["categories"] = [];
  for (const block of listed) {
    let category = categories.find((c) => c.title === block.category);
    if (!category) {
      category = { title: block.category!, blocks: [] };
      categories.push(category);
    }
    category.blocks.push({
      title: block.title,
      description: sentence(block.description),
      href: `${env.NEXT_PUBLIC_DOCS_LINK}/reference/blocks/${docsFolder(
        block.type
      )}/${block.type}`,
      plan: planFor(block),
    });
  }

  // Only Kite's own templates are listed by name, since the text of the
  // others is up to their authors.
  const templates = bundles && {
    count: bundles.length,
    official: bundles.filter((b) => b.official).length,
    top: bundles
      .filter((b) => b.official)
      .sort((a, b) => b.install_count - a.install_count)
      .slice(0, 6)
      .map((b) => ({
        id: b.id,
        name: b.listing?.name ?? b.name,
        description: b.listing?.description ?? b.description,
        installs: b.install_count,
      })),
  };

  return withApiData({
    blockCount: listed.length,
    eventCount: discordEventTypes.length,
    categories,
    paidBlocks: listed.filter((b) => b.premium_feature).map((b) => b.title),
    templates,
    bots: stats?.apps ?? null,
    free,
  });
};

// The folders of the block reference in kite-docs.
function docsFolder(type: string) {
  if (type.startsWith("control_")) return "controls";
  if (type.startsWith("entry_")) return "entries";
  if (type.startsWith("option_")) return "options";
  return "actions";
}

function sentence(text: string) {
  return /[.!?]$/.test(text) ? text : `${text}.`;
}

function list(items: string[]) {
  if (items.length < 3) return items.join(" and ");
  return `${items.slice(0, -1).join(", ")}, and ${items[items.length - 1]}`;
}

// Rounds down like the homepage, so the number never claims more than is true.
function floorThousands(value: number) {
  return `${(Math.floor(value / 1000) * 1000).toLocaleString("en-US")}+`;
}

export default function FeaturesPage({
  blockCount,
  eventCount,
  categories,
  paidBlocks,
  templates,
  bots,
  free,
}: Props) {
  const minutes = (n: number) =>
    n === 1 ? "every minute" : `every ${n} minutes`;

  const sections: {
    heading: string;
    body?: string;
    points: Point[];
    visual?: ReactNode;
    more?: { label: string; href: string };
  }[] = [
    {
      heading: "Start a flow from a command, a button, or an event",
      body: "Each command and event listener is a flow of blocks. It starts when one of these happens.",
      points: [
        {
          title: "Slash commands",
          icon: SquareSlashIcon,
          description:
            "Commands take arguments, like a user, role, channel, number, text, or file. You decide who can use each command, where it works, and how long the cooldown is.",
        },
        {
          title: "Buttons, select menus, and modals",
          icon: MousePointerClickIcon,
          description:
            "Add them to a message and run a flow when a member uses one. Modals ask members to fill in a form.",
        },
        {
          title: `${eventCount} Discord events`,
          icon: ZapIcon,
          description:
            "Run a flow without a command when messages are sent, edited, or deleted, when members join or leave, on reactions, boosts, direct messages, and audit log entries, and when your bot joins or leaves a server.",
        },
        {
          title: "Scheduled tasks",
          icon: ClockIcon,
          description:
            "Run a flow on a schedule, like a daily reminder." +
            (free
              ? ` The free plan runs them ${minutes(
                  free.scheduleMinutes
                )} at most${
                  free.fasterSchedule
                    ? `, and Kite ${free.fasterSchedule.plan} ${minutes(
                        free.fasterSchedule.minutes
                      )}`
                    : ""
                }.`
              : ""),
        },
        {
          title: "Webhooks",
          icon: WebhookIcon,
          description:
            "Start a flow from another website or service, like a store that tells your bot about a new order.",
        },
      ],
    },
    {
      heading: "Design messages and see them the way Discord shows them",
      body: "The message editor shows your message as it will look in Discord while you build it.",
      visual: <MessageMockup />,
      points: [
        {
          title: "Embeds and placeholders",
          icon: LayoutTemplateIcon,
          description:
            "Titles, colors, images, and fields. Placeholders fill in things like the member's name when the bot sends it.",
        },
        {
          title: "Discord's newer layouts",
          icon: LayoutGridIcon,
          description:
            "Containers, sections, and image galleries, next to classic embeds.",
        },
        {
          title: "Send it from anywhere",
          icon: SendIcon,
          description:
            "A command or event sends it, or you post it to a channel or forum from the dashboard.",
        },
      ],
    },
    {
      heading: "Conditions, loops, and stored variables for the logic",
      visual: <VariableMockup />,
      body: "Conditions and loops cost no credits, and neither does getting servers, channels, members, messages, or stored variables.",
      points: [
        {
          title: "Conditions",
          icon: SplitIcon,
          description:
            "If, comparison, role, user, and channel conditions run other blocks depending on who used the command and what they entered.",
        },
        {
          title: "Loops",
          icon: Repeat2Icon,
          description:
            "Run a set of blocks a number of times, and exit the loop early once you're done.",
        },
        {
          title: "Stored variables",
          icon: DatabaseIcon,
          description:
            "Keep points, counts, and settings for each member or server between runs. One block sets a value, adds to it, or subtracts from it.",
        },
        {
          title: "Math, random values, and tables",
          icon: CalculatorIcon,
          description:
            "Calculate Value does math, Generate Random Value picks a number or an item from a list, and Lookup Value finds a value in a table.",
        },
      ],
    },
    {
      heading: "AI that builds your bot, and AI your bot can use",
      points: [
        {
          title: "Flow AI",
          icon: SparklesIcon,
          description:
            "Describe a command and Flow AI, the AI assistant in the editor, adds and connects the blocks." +
            (free
              ? ` The free plan includes ${free.aiPrompts} prompts a month.`
              : ""),
          href: "/ai-discord-bot-maker",
        },
        {
          title: "AI blocks",
          icon: BrainIcon,
          description:
            "Ask AI answers questions, Search the Web looks things up, and AI Translate translates messages. AI Check, AI Choice, and AI Score sort messages, like spam or questions for your staff. They cost credits.",
          href: "/ai-discord-bot",
        },
        {
          title: "MCP server",
          icon: PlugIcon,
          description:
            "Build and change your bot from Claude, ChatGPT, Codex, Cursor, or VS Code.",
          href: "/discord-mcp-server",
        },
      ],
    },
    {
      heading: "Blocks for ER:LC, Roblox, and other services",
      points: [
        {
          title: "ER:LC",
          icon: SirenIcon,
          description:
            "Get the players and status of your private server and run commands on it, with your server key.",
          href: "/erlc-discord-bot",
        },
        {
          title: "Roblox",
          icon: Gamepad2Icon,
          description: "Look up a Roblox user by name or ID.",
        },
        {
          title: "Cookie API",
          icon: CookieIcon,
          description:
            "Make transcripts of channels, captchas, cards, and QR codes, and check Minecraft players.",
          href: "/cookie-api",
        },
        {
          title: "Any other website",
          icon: GlobeIcon,
          description:
            "Send API Request gets data from another website or sends it there. Discord API Request does anything Discord allows that has no block yet.",
        },
      ],
    },
    {
      heading: "Hosted 24/7 with your bot's own name and status",
      more: { label: "Compare every plan", href: "/pricing" },
      visual: (
        <div className="max-w-xs mx-auto">
          <ProfileMockup statuses={rotatingStatuses} />
        </div>
      ),
      points: [
        {
          title: "Always online",
          icon: ServerIcon,
          description:
            "Kite hosts your bot 24/7, whether it's used or not." +
            (free
              ? ` On the free plan, it can be in up to ${free.servers.toLocaleString(
                  "en-US"
                )} servers.`
              : ""),
          href: "/discord-bot-hosting",
        },
        {
          title: "Your own status",
          icon: CircleDotIcon,
          description:
            'Replace the default "Powered by Kite.onl" status with your own on every plan.',
        },
        {
          title: "Profile and voice on paid plans",
          icon: UserRoundIcon,
          description:
            "Kite Plus adds a rotating status, an animated avatar and banner, and audio in voice channels. Kite Premium adds a profile per server and the Send message as block.",
          href: "/pricing",
        },
        {
          title: "Collaborators",
          icon: UsersIcon,
          description: free
            ? `Up to ${free.collaborators} people, including you, can edit the same bot on the free plan.`
            : "Invite others to edit the same bot with you.",
        },
        { ...openSourcePoint, icon: CodeIcon },
      ],
    },
  ];

  const faqs: FAQProps[] = [
    {
      question: "How many blocks does Kite have?",
      answer: `${blockCount}. They're all listed on this page, grouped like in the editor, and each one has a page in the docs.`,
    },
    {
      question: "Does Kite have variables, conditions, and loops?",
      answer:
        "Yes. Stored variables keep data for each member or server between runs. Conditions and loops are blocks like any other, and they cost no credits.",
    },
    {
      question: "Can I make a leveling or economy bot with Kite?",
      answer:
        "Yes. Keep XP or coins for each member in a stored variable and add to it with a Set stored variable block. A Comparison Condition checks the value, for example to give a role at level 10. The Marketplace has no leveling or economy template right now, so you build it yourself or ask Flow AI to.",
    },
    {
      question: "Can I make a ticket bot with Kite?",
      answer:
        "Yes. The Marketplace has ticket templates from the community. The Modmail template by Kite lets members DM your bot to reach your staff.",
    },
    {
      question: "Which features are free?",
      answer: `Every block works on the free plan except ${list(
        paidBlocks
      )}. Most actions cost 1 credit${
        free
          ? `, and the free plan has ${free.credits.toLocaleString(
              "en-US"
            )} credits a month`
          : ""
      }.`,
    },
    {
      question: "Is Kite open source?",
      answer:
        "Yes. Kite's code is on GitHub under the AGPL-3.0 license, so you can see how your bot runs or host Kite yourself.",
    },
  ];

  const proof = [
    { value: blockCount.toString(), label: "blocks" },
    { value: eventCount.toString(), label: "Discord events" },
    templates && { value: templates.count.toString(), label: "templates" },
    bots && { value: floorThousands(bots), label: "bots built" },
  ].filter((p) => !!p);

  return (
    <HomeLayout
      title="Kite features: every block of the free Discord bot maker"
      description={`Kite has ${blockCount} blocks for Discord bot commands and events, with conditions, loops, stored variables, AI, and ${
        templates ? `${templates.count} ` : ""
      }templates. See what's free and what each plan adds.`}
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="container max-w-4xl py-16 md:py-24 text-center">
        <h1 className="text-4xl md:text-5xl font-bold">
          What your Discord bot can do with Kite
        </h1>
        <p className="text-xl text-muted-foreground mt-6">
          Kite has {blockCount} blocks that you connect into commands and event
          listeners. They send messages, manage members and roles, use AI, and
          keep data between runs. Your bot is online 24/7 on the free plan.
        </p>
        <div className="mt-8 flex flex-col md:flex-row justify-center gap-4">
          <Button className="md:w-48" asChild>
            <Link href="/apps">Get started for free</Link>
          </Button>
          <Button variant="outline" className="md:w-48" asChild>
            <Link href="/marketplace">Browse templates</Link>
          </Button>
        </div>
        <a
          href="#blocks"
          className="inline-block mt-6 text-sm text-muted-foreground underline hover:text-foreground"
        >
          Or jump to all {blockCount} blocks
        </a>
      </section>

      <section className="border-y bg-muted/30">
        <div className="container grid grid-cols-2 lg:grid-cols-4 gap-x-4 gap-y-8 py-10">
          {proof.map((item) => (
            <div key={item.label} className="text-center">
              <div
                className="text-2xl sm:text-3xl md:text-4xl font-bold"
                data-indexnow-ignore
              >
                {item.value}
              </div>
              <div className="text-sm md:text-base text-muted-foreground mt-1">
                {item.label}
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="container py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold">
          Start with the bot you need
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-2xl">
          Most people start with one of these. Each one links to a template you
          can add in a few clicks, or a page that shows how to build it.
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-10">
          {starters.map((starter) => {
            const Icon = starter.icon!;
            return (
              <Link
                key={starter.title}
                href={starter.href!}
                className="group rounded-xl border bg-card p-5 hover:border-primary transition-colors"
              >
                <div className="flex items-center gap-3">
                  <span className="flex h-9 w-9 flex-none items-center justify-center rounded-lg bg-primary/10 text-primary">
                    <Icon className="h-5 w-5" />
                  </span>
                  <span className="font-semibold text-lg">{starter.title}</span>
                  <ArrowRightIcon className="ml-auto h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                </div>
                <p className="text-muted-foreground mt-3">
                  {starter.description}
                </p>
              </Link>
            );
          })}
        </div>
      </section>

      {sections.slice(0, 4).map((section) => (
        <PointsSection key={section.heading} {...section} />
      ))}

      {templates && (
        <PointsSection
          heading="Templates to start from instead of an empty bot"
          body={`The Marketplace has ${templates.count} templates, ${
            templates.official
          } made by Kite and ${
            templates.count - templates.official
          } from the community. Adding one gives you normal commands and event listeners that you can change. These templates by Kite are added the most.`}
          more={{
            label: `Browse all ${templates.count} templates`,
            href: "/marketplace",
          }}
          points={templates.top.map((t) => ({
            title: t.name,
            description: `${sentence(
              t.description
            )} Added ${t.installs.toLocaleString("en-US")} times.`,
            href: `/marketplace/${t.id}`,
          }))}
        />
      )}

      {sections.slice(4).map((section) => (
        <PointsSection key={section.heading} {...section} />
      ))}

      <section id="blocks" className="container py-12 md:py-16 scroll-mt-20">
        <h2 className="text-3xl md:text-4xl font-bold">
          All {blockCount} blocks
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-2xl">
          Every block of the editor, in the groups of its block list. Hover one
          to see what it does, or click it for its page in the docs.
        </p>
        <div className="mt-10 columns-1 md:columns-2 lg:columns-3 gap-4">
          {categories.map((category) => (
            <div
              key={category.title}
              className="break-inside-avoid mb-4 rounded-xl border bg-card p-5"
            >
              <h3 className="font-semibold flex items-baseline justify-between gap-3">
                {category.title}
                <span className="text-sm font-normal text-muted-foreground">
                  {category.blocks.length}
                </span>
              </h3>
              <ul className="mt-3 flex flex-wrap gap-2">
                {category.blocks.map((block) => (
                  <li key={block.title}>
                    <a
                      href={block.href}
                      title={block.description}
                      className="inline-flex items-center gap-1.5 rounded-md border bg-muted/40 px-2 py-1 text-sm hover:border-primary hover:text-primary transition-colors"
                    >
                      {block.title}
                      {/* The tooltip is only for the mouse. This way screen
                          readers and crawlers get the description too. */}
                      <span className="sr-only">: {block.description}</span>
                      {block.plan && (
                        <span className="text-xs font-medium text-primary">
                          {block.plan}
                        </span>
                      )}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </section>

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

// With a visual, the text goes on the left and the visual on the right, like
// the landing pages.
function PointsSection({
  heading,
  body,
  points,
  visual,
  more,
}: {
  heading: string;
  body?: string;
  points: Point[];
  visual?: ReactNode;
  // A page that tells more, linked below the points.
  more?: { label: string; href: string };
}) {
  const moreLink = more && (
    <Link
      href={more.href}
      className="mt-6 inline-flex items-center gap-1 font-medium hover:text-primary transition-colors"
    >
      {more.label}
      <ArrowRightIcon className="w-4 h-4" />
    </Link>
  );
  const header = (
    <>
      <h2 className="text-3xl md:text-4xl font-bold">{heading}</h2>
      {body && (
        <p className="text-lg text-muted-foreground mt-4 max-w-2xl">{body}</p>
      )}
    </>
  );
  if (visual) {
    return (
      <section className="container py-12 md:py-16 grid grid-cols-1 lg:grid-cols-2 gap-10 lg:gap-16 items-center">
        <div>
          {header}
          <PointList points={points} className="sm:grid-cols-2" />
          {moreLink}
        </div>
        <div className="w-full max-w-lg mx-auto">{visual}</div>
      </section>
    );
  }
  return (
    <section className="container py-12 md:py-16">
      {header}
      <PointList points={points} className="md:grid-cols-2 lg:grid-cols-3" />
      {moreLink}
    </section>
  );
}

function PointList({
  points,
  className,
}: {
  points: Point[];
  className: string;
}) {
  return (
    <ul className={cn("grid grid-cols-1 gap-x-10 mt-10", className)}>
      {points.map((point) => (
        <li key={point.title} className="border-t py-5">
          <h3 className="font-semibold text-lg flex items-center gap-3">
            {point.icon && (
              <point.icon className="h-5 w-5 flex-none text-primary" />
            )}
            {point.href ? (
              <Link
                href={point.href}
                className="inline-flex items-center gap-1 hover:text-primary transition-colors"
              >
                {point.title}
                <ArrowRightIcon className="w-4 h-4" />
              </Link>
            ) : (
              point.title
            )}
          </h3>
          <p className="text-muted-foreground mt-2">{point.description}</p>
        </li>
      ))}
    </ul>
  );
}
