import MCPClientSetup from "@/components/app/MCPClientSetup";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import DiscordIcon from "@/components/icons/DiscordIcon";
import DiscordReplyMockup from "@/components/landing/DiscordReplyMockup";
import MCPSessionMockup, {
  MCPSession,
} from "@/components/landing/MCPSessionMockup";
import ShowcaseVisual from "@/components/landing/ShowcaseVisual";
import { welcomeFlow } from "@/components/landing/flows";
import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import { faqJsonLd } from "@/lib/seo";
import { cn } from "@/lib/utils";
import {
  BanIcon,
  CheckIcon,
  CodeIcon,
  HistoryIcon,
  LayoutTemplateIcon,
  PencilLineIcon,
  PlugIcon,
  SparklesIcon,
  UserRoundIcon,
  WrenchIcon,
  XIcon,
} from "lucide-react";
import Link from "next/link";
import { ComponentType, ReactNode } from "react";
import { toast } from "sonner";

const session: MCPSession = {
  prompt:
    "When someone joins, give them the Member role and welcome them in #welcome.",
  calls: [
    { tool: "list_roles", result: "Found @Member" },
    { tool: "list_channels", result: "Found #welcome" },
    { tool: "get_blocks", result: "Looked up 3 blocks" },
    { tool: "create_event_listener", result: "Created Welcome new members" },
  ],
  answer:
    "Added an event listener that gives new members @Member and welcomes them in #welcome. It's live in your server.",
};

const clients = [
  "Claude",
  "ChatGPT",
  "Claude Code",
  "Codex",
  "Cursor",
  "VS Code",
];

const steps = [
  {
    title: "Add Kite to Your Client",
    description:
      "Paste Kite's URL into Claude, ChatGPT, Claude Code or Codex. Cursor and a few others use an API token instead.",
  },
  {
    title: "Sign In and Pick Its Access",
    description:
      "Choose which apps the client can use, and whether it can change them or only read them.",
  },
  {
    title: "Describe Your Bot",
    description:
      "Ask for a command or an event in plain words. The client builds the flow, Kite checks it, and it goes live in your server.",
  },
];

const examples: {
  icon: ComponentType<{ className?: string }>;
  label: string;
  color: string;
  prompt: string;
  description: string;
}[] = [
  {
    icon: SparklesIcon,
    label: "Build",
    color: "bg-primary/10 text-primary",
    prompt:
      "Add a /poll command with a question and up to four options, and buttons to vote.",
    description:
      "Creates the command, the message with its buttons and a flow for each button.",
  },
  {
    icon: WrenchIcon,
    label: "Fix",
    color: "bg-rose-500/10 text-rose-500",
    prompt: "Why does my /ticket command fail? Check the logs and fix it.",
    description:
      "Reads your app's logs, finds the block that failed and fixes its settings.",
  },
  {
    icon: PencilLineIcon,
    label: "Change",
    color: "bg-sky-500/10 text-sky-500",
    prompt: "Make all of my moderation commands log to #mod-log.",
    description:
      "Looks up your channels and changes every flow that needs it, in one go.",
  },
  {
    icon: DiscordIcon,
    label: "Discord API",
    color: "bg-[#5865f2]/10 text-[#5865f2]",
    prompt: "Create an #announcements channel that only @Staff can post in.",
    description:
      "With an API token, it sends the requests to Discord as your bot, for things that don't need a flow.",
  },
];

const comparison = [
  {
    feature: "Hosting",
    code: "A server you pay for and keep running",
    kite: "Free and 24/7 on Kite's servers",
  },
  {
    feature: "Changes",
    code: "Edit the code and deploy it again",
    kite: "Ask the client, or edit the flow yourself",
  },
  {
    feature: "Buttons, modals and variables",
    code: "Code the client writes and you debug",
    kite: "Built-in blocks that Kite checks",
  },
  {
    feature: "When it breaks",
    code: "Logs on your server, if you set them up",
    kite: "Logs in the dashboard that the client can read",
  },
];

const extras = [
  {
    icon: LayoutTemplateIcon,
    title: "Templates to Start From",
    description:
      "Add verification, modmail or giveaways from the Marketplace and ask the client to change it for your server.",
  },
  {
    icon: UserRoundIcon,
    title: "Your Own Bot",
    description:
      "The bot is your own Discord application, with your name, avatar and status, not Kite's.",
  },
  {
    icon: CodeIcon,
    title: "Open Source",
    description:
      "Kite's code is on GitHub. You can see how your bot runs, or host Kite yourself.",
  },
];

const controls = [
  {
    title: "Read-only access",
    description:
      "The client can look at your apps and logs without changing anything.",
  },
  {
    title: "Only some apps",
    description:
      "Give a client only the apps you pick. It doesn't see the rest.",
  },
  {
    title: "Every change logged",
    description:
      "Requests to Discord with an API token show up in your app's logs and the server's audit log with the token's name.",
  },
];

const faqs: FAQProps[] = [
  {
    question: "What is an MCP server?",
    answer:
      "MCP, the Model Context Protocol, is a standard way for AI clients to use tools of other apps. Kite's MCP server gives clients tools to read and build your Discord bot, like listing your commands or adding a block to a flow.",
  },
  {
    question: "Which AI clients work with Kite?",
    answer:
      "Claude, ChatGPT, Claude Code, Codex and VS Code connect by signing in to Kite. Cursor, OpenCode, Antigravity, Gemini CLI, Devin Desktop and other clients that connect to remote MCP servers use an API token instead.",
  },
  {
    question: "Does it cost anything?",
    answer:
      "No. Kite hosts your bot 24/7 for free, and the MCP server works on every plan without using credits or flow AI prompts. The flows it builds use credits when they run, like any other flow, and the free plan has 10,000 a month.",
  },
  {
    question: "Why not have Claude Code or Codex write a discord.js bot?",
    answer:
      "You can, but then you host the bot yourself and keep it running, and only someone who reads the code can change it. With Kite, the bot runs on Kite's servers for free, and anyone on your team can open what the client built in the editor and change it.",
  },
  {
    question: "Is it safe to let an AI change my bot?",
    answer:
      "You choose what each client can do, and every change can be undone from the history. Kite tells the client to only delete or disable things you asked for. If you use an API token, keep it private, because anyone who has it can control your bot.",
  },
  {
    question: "Do I need to host the bot myself?",
    answer:
      "No. Kite runs your bot 24/7. The AI client only builds and changes it, so the bot keeps running when you close the client.",
  },
];

function Highlight({ children }: { children: ReactNode }) {
  return (
    <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
      {children}
    </span>
  );
}

function copy(text: string) {
  navigator.clipboard
    .writeText(text)
    .then(() => toast.success("Copied to clipboard!"))
    .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
}

export default function DiscordMCPServerPage() {
  return (
    <HomeLayout
      title="Discord MCP Server for Claude, ChatGPT, Claude Code and Codex"
      description="Connect Claude, ChatGPT, Claude Code or Codex to Kite's MCP server and let AI build the commands and events of your Discord bot. Kite hosts the bot for you."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="relative container grid grid-cols-1 lg:grid-cols-2 gap-12 items-center py-16 md:py-24">
        <div className="z-10 text-center lg:text-start">
          <h1 className="text-4xl md:text-6xl font-bold">
            Build Your Discord Bot From{" "}
            <span className="bg-gradient-to-r from-[#f9ad15] to-primary text-transparent bg-clip-text">
              Claude
            </span>{" "}
            or{" "}
            <span className="bg-gradient-to-r from-[#34d399] to-[#10a37f] text-transparent bg-clip-text">
              ChatGPT
            </span>
          </h1>
          <p className="text-xl text-muted-foreground mt-6">
            Connect your AI client to Kite, describe what your bot should do,
            and it builds the commands, event listeners and messages. Kite runs
            the bot 24/7 for free, so there&apos;s nothing to deploy or host.
          </p>
          <div className="mt-8 flex flex-col md:flex-row justify-center lg:justify-start gap-4">
            <Button className="md:w-48" asChild>
              <Link href="/apps">Get Started for Free</Link>
            </Button>
            <Button variant="outline" className="md:w-48" asChild>
              <a href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/mcp`}>
                Read the MCP Guide
              </a>
            </Button>
          </div>
          <div className="mt-8 flex flex-wrap items-center justify-center lg:justify-start gap-2 text-sm text-muted-foreground">
            <span>Works with</span>
            {clients.map((client) => (
              <span
                key={client}
                className="rounded-md border bg-muted/50 px-2 py-0.5 font-medium text-foreground"
              >
                {client}
              </span>
            ))}
            <span>and more</span>
          </div>
        </div>

        <div className="z-10 min-w-0">
          <MCPSessionMockup session={session} />
          <div className="ml-auto mt-3 max-w-sm">
            <div className="flex items-center gap-2 pl-6 pb-3 text-sm text-muted-foreground">
              <div className="h-6 border-l-2 border-dashed" />
              <DiscordIcon className="size-4 text-[#5865f2]" />A minute later,
              in your server
            </div>
            <DiscordReplyMockup
              event="Alex joined the server."
              content="Welcome to the server, @Alex!"
            />
          </div>
        </div>

        {/* Shadow effect */}
        <div className="shadow"></div>
      </section>

      <section
        id="setup"
        className="container grid grid-cols-1 lg:grid-cols-[2fr_3fr] gap-12 items-center py-16 md:py-24 scroll-mt-16"
      >
        <div>
          <h2 className="text-3xl md:text-4xl font-bold">
            Set It Up in <Highlight>Two Minutes</Highlight>
          </h2>
          <ol className="relative mt-8 space-y-8">
            {/* Connects the numbers of the steps. */}
            <div className="absolute left-5 top-5 bottom-5 border-l-2 border-dashed" />
            {steps.map((step, i) => (
              <li key={step.title} className="relative flex gap-4">
                <div className="flex-none size-10 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  {i + 1}
                </div>
                <div>
                  <h3 className="text-lg font-semibold">{step.title}</h3>
                  <p className="text-muted-foreground mt-1">
                    {step.description}
                  </p>
                </div>
              </li>
            ))}
          </ol>
        </div>
        <div className="min-w-0 rounded-xl border bg-card p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-3 font-semibold">
            <PlugIcon className="size-4 text-primary" />
            Connect your client
          </div>
          <MCPClientSetup
            url={`${env.NEXT_PUBLIC_SITE_URL}/api/mcp`}
            copy={copy}
          />
          <p className="mt-3 text-sm text-muted-foreground">
            For clients that need a token, Kite shows their setup when you
            create one under API tokens.
          </p>
        </div>
      </section>

      <section className="container py-16 md:py-24">
        <div className="text-center max-w-3xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold">
            Ask for <Highlight>Anything</Highlight> Your Bot Should Do
          </h2>
          <p className="text-lg text-muted-foreground mt-4">
            The client sees your commands, logs, channels and roles, and looks
            up how each block works before it uses it.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-12 max-w-5xl mx-auto">
          {examples.map((example) => (
            <div
              key={example.label}
              className="rounded-xl border bg-card p-6 shadow-sm"
            >
              <div
                className={cn(
                  "inline-flex items-center gap-2 rounded-full px-3 py-1 text-sm font-medium",
                  example.color
                )}
              >
                <example.icon className="size-4" />
                {example.label}
              </div>
              <div className="mt-4 rounded-lg rounded-tl-sm bg-muted px-4 py-3 font-medium">
                {example.prompt}
              </div>
              <p className="mt-3 text-muted-foreground">
                {example.description}
              </p>
            </div>
          ))}
        </div>
        <div className="mt-8 flex flex-col md:flex-row justify-center gap-4 md:gap-10 text-muted-foreground">
          <div className="flex items-center gap-2">
            <HistoryIcon className="size-5 flex-none text-primary" />
            Every change is kept in the history, so you can restore it.
          </div>
          <div className="flex items-center gap-2">
            <SparklesIcon className="size-5 flex-none text-primary" />
            Uses your client&apos;s model, not Kite&apos;s AI prompts.
          </div>
        </div>
      </section>

      <section className="container py-16 md:py-24">
        <div className="text-center max-w-3xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold">
            Why Not Have It <Highlight>Write Code</Highlight>?
          </h2>
          <p className="text-lg text-muted-foreground mt-4">
            Claude Code and Codex can write a discord.js or discord.py bot too.
            Then you need a server that runs it, and every change goes through
            the code. On Kite, the client builds flows that Kite runs, and you
            can open them in the editor and change them yourself.
          </p>
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mt-12 items-center">
          <div className="rounded-xl border bg-card shadow-sm overflow-hidden">
            <div className="grid grid-cols-2 border-b bg-muted/50 text-sm font-semibold">
              <div className="px-5 py-3 text-muted-foreground">
                A bot in code
              </div>
              <div className="px-5 py-3 text-primary">Kite</div>
            </div>
            {comparison.map((row) => (
              <div key={row.feature} className="border-b last:border-b-0">
                <div className="px-5 pt-4 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {row.feature}
                </div>
                <div className="grid grid-cols-2">
                  <div className="flex gap-2 px-5 pt-2 pb-4 text-muted-foreground">
                    <XIcon className="size-4 mt-1 flex-none text-rose-500" />
                    {row.code}
                  </div>
                  <div className="flex gap-2 px-5 pt-2 pb-4 font-medium">
                    <CheckIcon className="size-4 mt-1 flex-none text-emerald-500" />
                    {row.kite}
                  </div>
                </div>
              </div>
            ))}
          </div>
          <div className="min-w-0">
            <div className="mb-3 flex items-center gap-2 text-sm text-muted-foreground">
              <PencilLineIcon className="size-4 text-primary" />
              What the client built, open in Kite&apos;s editor
            </div>
            <ShowcaseVisual
              showcase={{
                flow: welcomeFlow,
                event: "Alex joined the server.",
                reply: "Welcome to the server, @Alex!",
              }}
            />
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-12">
          {extras.map((extra) => (
            <div key={extra.title} className="flex gap-4">
              <div className="flex-none size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
                <extra.icon className="size-5" />
              </div>
              <div>
                <h3 className="font-semibold">{extra.title}</h3>
                <p className="text-muted-foreground mt-1">
                  {extra.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="container grid grid-cols-1 lg:grid-cols-2 gap-12 items-center py-16 md:py-24">
        <div>
          <h2 className="text-3xl md:text-4xl font-bold">
            You Decide What It <Highlight>Can Touch</Highlight>
          </h2>
          <p className="text-lg text-muted-foreground mt-4">
            Each client has its own access, so it only gets what you give it.
            Disconnect it under API tokens and it loses access right away.
          </p>
          <ul className="mt-8 space-y-4">
            {controls.map((control) => (
              <li key={control.title} className="flex gap-3">
                <CheckIcon className="size-5 mt-0.5 flex-none text-primary" />
                <span>
                  <span className="font-semibold">{control.title}.</span>{" "}
                  <span className="text-muted-foreground">
                    {control.description}
                  </span>
                </span>
              </li>
            ))}
          </ul>
        </div>
        <ConnectionMockup />
      </section>

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

// A simplified connection from the consent page, with what it can do.
function ConnectionMockup() {
  const settings = [
    { label: "Apps access", value: "Full access" },
    { label: "Which apps", value: "Ticket Bot, Mod Bot" },
  ];
  const blocked = [
    "Mass bans",
    "Bulk deletes",
    "Pruning members",
    "Leaving servers",
  ];

  return (
    <div className="rounded-xl border bg-card p-6 shadow-sm">
      <div className="flex items-center gap-3">
        <div className="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
          <PlugIcon className="size-5" />
        </div>
        <div>
          <div className="font-semibold">Claude</div>
          <div className="text-sm text-muted-foreground">
            Connected from claude.ai · Used 3 minutes ago
          </div>
        </div>
      </div>
      <div className="grid sm:grid-cols-2 gap-3 mt-6">
        {settings.map((setting) => (
          <div key={setting.label}>
            <div className="text-sm font-medium">{setting.label}</div>
            <div className="mt-1.5 rounded-md border bg-background px-3 py-2 text-sm truncate">
              {setting.value}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 border-t pt-5">
        <div className="flex items-center gap-2 text-sm font-medium">
          <BanIcon className="size-4 text-rose-500" />
          Always blocked for clients
        </div>
        <div className="flex flex-wrap gap-2 mt-3">
          {blocked.map((item) => (
            <span
              key={item}
              className="rounded-full bg-rose-500/10 px-3 py-1 text-sm text-rose-500"
            >
              {item}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
