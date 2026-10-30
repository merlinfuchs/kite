import env from "@/lib/env/client";
import { ArrowRightIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";

// Simplified versions of what they show, drawn with the site's colors like the
// Ask AI panel.

function MessageMockup() {
  return (
    <div className="rounded-xl border bg-card p-4 shadow-sm text-sm">
      <div className="rounded border-l-4 border-primary bg-muted px-3 py-2">
        <div className="font-semibold">📜 Server Rules</div>
        <div className="mt-1 text-muted-foreground">
          Be kind, keep it on topic and no spam. Click below to get access to
          the server.
        </div>
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        <span className="rounded bg-green-600 px-3 py-1 text-xs font-medium text-white">
          ✅ Accept rules
        </span>
        <span className="rounded bg-muted px-3 py-1 text-xs font-medium">
          🎫 Open a ticket
        </span>
      </div>
    </div>
  );
}

const points = [
  { name: "Alex", value: 1240 },
  { name: "Sam", value: 980 },
  { name: "Jordan", value: 455 },
];

function VariableMockup() {
  const max = Math.max(...points.map((p) => p.value));
  return (
    <div className="rounded-xl border bg-card p-4 shadow-sm text-sm">
      <div className="flex items-center justify-between">
        <span className="font-mono font-medium">points</span>
        <span className="text-xs text-muted-foreground">per member</span>
      </div>
      <div className="mt-3 space-y-2">
        {points.map((p) => (
          <div key={p.name} className="flex items-center gap-3">
            <span className="w-14 flex-none">{p.name}</span>
            <div className="h-2 flex-1 rounded-full bg-muted">
              <div
                className="h-2 rounded-full bg-primary"
                style={{ width: `${(p.value / max) * 100}%` }}
              />
            </div>
            <span className="w-12 flex-none text-right tabular-nums text-muted-foreground">
              {p.value.toLocaleString("en-US")}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

const templates = [
  { name: "Giveaways", description: "Enter button, random winners" },
  { name: "Modmail", description: "Members DM your staff" },
  { name: "Verification", description: "A role for clicking Verify" },
];

function TemplateMockup() {
  return (
    <div className="space-y-2 rounded-xl border bg-card p-4 shadow-sm text-sm">
      {templates.map((t) => (
        <div key={t.name} className="flex items-center justify-between gap-3">
          <div className="min-w-0">
            <div className="font-medium">{t.name}</div>
            <div className="truncate text-xs text-muted-foreground">
              {t.description}
            </div>
          </div>
          <span className="flex-none rounded-md border px-2.5 py-1 text-xs font-medium">
            Add
          </span>
        </div>
      ))}
    </div>
  );
}

const features: {
  title: string;
  description: string;
  href?: string;
  mockup: ReactNode;
}[] = [
  {
    title: "Message Editor",
    description:
      "Design messages with embeds, buttons and select menus, and see them the way Discord shows them.",
    mockup: <MessageMockup />,
  },
  {
    title: "Stored Variables",
    description:
      "Keep points, counts and settings for each member or server, even across restarts.",
    mockup: <VariableMockup />,
  },
  {
    title: "Templates",
    description:
      "Add one of 50+ ready-made bots from the Marketplace, then change it to fit your server.",
    href: "/marketplace",
    mockup: <TemplateMockup />,
  },
];

// The rest of what every plan has, smaller and with a hint of each instead of a
// whole mockup.
const extras: {
  title: string;
  description: string;
  href?: string;
  visual: ReactNode;
}[] = [
  {
    title: "Online 24/7",
    description:
      "Kite hosts the bot, so it stays up when your computer is off.",
    visual: (
      <div className="flex gap-0.5">
        {Array.from({ length: 14 }, (_, i) => (
          <span key={i} className="h-6 w-1.5 rounded-sm bg-green-500" />
        ))}
      </div>
    ),
  },
  {
    title: "Any HTTP API and Roblox",
    description:
      "Call any HTTP API from a flow, or look up Roblox players with a ready block.",
    visual: (
      <code className="rounded-md bg-muted px-2 py-1 text-xs">
        GET /weather <span className="text-green-600">200</span>
      </code>
    ),
  },
  {
    title: "Build Together",
    description: "Invite collaborators who edit the same bot with you.",
    visual: (
      <div className="flex -space-x-2">
        {["🦊", "🐼", "🐸"].map((emoji) => (
          <span
            key={emoji}
            className="flex h-7 w-7 items-center justify-center rounded-full border-2 border-background bg-muted text-sm"
          >
            {emoji}
          </span>
        ))}
      </div>
    ),
  },
  {
    title: "Open Source",
    description: "See how Kite works, or host it on your own server.",
    href: env.NEXT_PUBLIC_GITHUB_LINK,
    visual: (
      <code className="rounded-md bg-muted px-2 py-1 text-xs">
        {env.NEXT_PUBLIC_GITHUB_LINK.replace(/^https?:\/\//, "")}
      </code>
    ),
  },
];

export default function HomeFeaturesSection() {
  return (
    <section id="features" className="scroll-mt-16 container py-16 sm:py-24">
      <h2 className="text-3xl md:text-4xl font-bold">More Than Commands</h2>
      <p className="text-lg md:text-xl text-muted-foreground mt-4 max-w-2xl">
        Messages with buttons, data that sticks around and bots you can add in
        one click, all on the free plan.
      </p>

      {/* The mockups, titles and descriptions each line up in a row. */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-x-8 gap-y-10 md:gap-y-0 mt-8 md:mt-12">
        {features.map((feature) => (
          <div
            key={feature.title}
            className="grid md:row-span-3 md:grid-rows-subgrid"
          >
            <div className="[&>*]:h-full">{feature.mockup}</div>
            <h3 className="text-lg font-semibold mt-5">
              {feature.href ? (
                <Link
                  href={feature.href}
                  className="inline-flex items-center gap-1 hover:text-primary transition-colors"
                >
                  {feature.title}
                  <ArrowRightIcon className="w-4 h-4" />
                </Link>
              ) : (
                feature.title
              )}
            </h3>
            <p className="text-muted-foreground mt-1">{feature.description}</p>
          </div>
        ))}
      </div>

      <div className="mt-12 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8 border-t pt-10">
        {extras.map((extra) => (
          <div key={extra.title}>
            <div className="flex h-7 items-center">{extra.visual}</div>
            <h3 className="mt-3 font-semibold">
              {extra.href ? (
                <a
                  href={extra.href}
                  target="_blank"
                  className="inline-flex items-center gap-1 hover:text-primary transition-colors"
                >
                  {extra.title}
                  <ArrowRightIcon className="w-4 h-4" />
                </a>
              ) : (
                extra.title
              )}
            </h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {extra.description}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
