import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import env from "@/lib/env/client";
import { ArrowRightIcon } from "lucide-react";
import Link from "next/link";

// The free limits are typed by hand here and on each comparison page. Change
// both when a bot maker changes its plans.
const botMakers: {
  name: string;
  href: string;
  linkLabel: string;
  summary: string;
  freePlan: string;
  paidPlans?: string;
  bestFor: string;
  lastChecked?: string;
}[] = [
  {
    name: "Kite",
    href: "/",
    linkLabel: "See how Kite works",
    summary:
      "You build commands and event listeners by connecting blocks in a visual editor, or describe them to Flow AI, the AI assistant that builds your bot. You can also start from over 100 templates, like moderation, giveaways, and welcome messages, and change every block. Kite hosts the bot 24/7 and is open source.",
    freePlan:
      "15 commands, 10 event listeners, and 10,000 credits a month per bot, on up to 100 servers. You can make up to 10 bots and set your own status for free. Bots stay online even when nobody uses them.",
    paidPlans: "From $2.99 a month per bot.",
    bestFor:
      "Your own bot with your own commands, without paying or writing code.",
  },
  {
    name: "BotGhost",
    href: "/alternatives/botghost",
    linkLabel: "Kite vs. BotGhost",
    summary:
      "A no-code bot maker with more than 35 ready-made modules and a marketplace, plus blocks for your own commands.",
    freePlan:
      'One bot online at a time, on up to 5 servers, with 3 custom commands and 3 custom events. 11 modules, like tickets and giveaways, are free and don\'t count toward those limits. You can change their settings, but changing their blocks needs Premium. The status says "Made with BotGhost", and the bot is switched off after 2 days without activity.',
    paidPlans: "$12.50 a month or $75 a year per bot.",
    bestFor:
      "Switching on ready-made modules like tickets, giveaways, and reaction roles instead of building commands.",
    lastChecked: "November 2026",
  },
  {
    name: "Bot Designer for Discord (BDFD)",
    href: "/alternatives/bdfd",
    linkLabel: "Kite vs. BDFD",
    summary:
      "A bot builder for the web, Android, and iOS. You write commands in BDScript, its own scripting language.",
    freePlan:
      "Free, but your bot runs on hosting time that you earn with a button or by watching ads.",
    bestFor:
      "Building your bot on your phone, or if you already know BDScript.",
    lastChecked: "October 2026",
  },
  {
    name: "inventor.bot",
    href: "/alternatives/inventor",
    linkLabel: "Kite vs. inventor.bot",
    summary:
      "A visual bot builder, formerly inventor.gg, that you can also edit with your own AI agent. It has plugins and a built-in database.",
    freePlan:
      "2 bots, each on up to 15 servers, with 15 commands and events combined. The status shows an invt.bot ad, and a bot goes offline after 14 days without an event.",
    bestFor:
      "Building your bot with an AI agent of your choice, or storing data in a built-in database.",
    lastChecked: "October 2026",
  },
  {
    name: "SCNX",
    href: "/alternatives/scnx",
    linkLabel: "Kite vs. SCNX",
    summary:
      "A bot maker built around ready-made modules, like social media notifications, tickets, and modmail.",
    freePlan:
      'One server per bot and 10 custom triggers. Embeds say "Powered by scnx.xyz", and you watch an ad every 7 to 14 days to keep the bot online.',
    bestFor:
      "Dozens of modules you switch on, for a bot that only needs one server.",
    lastChecked: "October 2026",
  },
  {
    name: "FromFlow",
    href: "/alternatives/fromflow",
    linkLabel: "Kite vs. FromFlow",
    summary:
      "A visual bot builder with an AI agent and ready-made apps for tickets, leveling, music, and more.",
    freePlan: "No free plan. The 3-day trial needs a $1 card check.",
    paidPlans: "$4.99 a month, or a one-time payment for lifetime access.",
    bestFor: "Ready-made apps, or paying once instead of every month.",
    lastChecked: "October 2026",
  },
  {
    name: "BotWiz",
    href: "/alternatives/botwiz",
    linkLabel: "Kite vs. BotWiz",
    summary:
      "A block-based bot builder that can also host bots you wrote yourself in Node.js or Python.",
    freePlan: "No free plan. After a 7-day trial it needs a subscription.",
    bestFor: "Hosting your own Node.js or Python bot in the same place.",
    lastChecked: "October 2026",
  },
  {
    name: "VibeBot",
    href: "/alternatives/vibebot",
    linkLabel: "Kite vs. VibeBot",
    summary:
      "An AI bot maker. You describe what the bot should do, and the AI writes the code.",
    freePlan:
      "A shared bot named VibeBot. Your own bot with your own name and avatar needs a paid plan.",
    paidPlans: "$5 a month for your own bot, or $10 with the AI builder.",
    bestFor:
      "AI that writes real code, a music player, or exporting your bot's code.",
    lastChecked: "October 2026",
  },
];

const jsonLd = [
  {
    "@context": "https://schema.org",
    "@type": "ItemList",
    name: "The best free Discord bot makers in 2026",
    itemListElement: botMakers.map((b, i) => ({
      "@type": "ListItem",
      position: i + 1,
      name: b.name,
      url: env.NEXT_PUBLIC_SITE_URL + b.href,
    })),
  },
];

export default function AlternativesPage() {
  return (
    <HomeLayout
      title="Best free Discord bot makers in 2026"
      description="Eight Discord bot builders compared, including Kite, BotGhost, BDFD, and inventor.bot. See which ones are free, what each free plan includes, and when each one is the better choice."
      jsonLd={jsonLd}
    >
      <section className="container py-16 md:py-24 text-center max-w-3xl">
        <h1 className="text-4xl md:text-5xl font-bold">
          The best free Discord bot makers in 2026
        </h1>
        <p className="text-xl text-muted-foreground mt-6">
          See what each free plan includes and when each bot maker is the better
          choice. We checked every plan in October or November 2026.
        </p>
      </section>

      <section className="container pb-12 md:pb-16">
        <div className="max-w-3xl mx-auto space-y-6">
          {botMakers.map((b) => (
            <div
              key={b.name}
              className="rounded-lg border bg-muted/50 p-6 md:p-8"
            >
              <h2 className="text-2xl font-semibold">{b.name}</h2>
              <p className="text-muted-foreground mt-2">{b.summary}</p>
              <dl className="mt-5 space-y-3">
                <div>
                  <dt className="font-medium">Free plan</dt>
                  <dd className="text-muted-foreground">{b.freePlan}</dd>
                </div>
                {b.paidPlans && (
                  <div>
                    <dt className="font-medium">Paid plans</dt>
                    <dd className="text-muted-foreground">{b.paidPlans}</dd>
                  </div>
                )}
                <div>
                  <dt className="font-medium">Best for</dt>
                  <dd className="text-muted-foreground">{b.bestFor}</dd>
                </div>
              </dl>
              <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2 mt-6">
                <Link
                  href={b.href}
                  className="group flex items-center gap-2 font-medium text-primary"
                >
                  {b.linkLabel}
                  <ArrowRightIcon className="w-4 h-4 transition-transform group-hover:translate-x-1" />
                </Link>
                {b.lastChecked && (
                  <span className="text-sm text-muted-foreground">
                    Checked {b.lastChecked}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      </section>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
