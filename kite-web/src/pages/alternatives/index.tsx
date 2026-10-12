import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { ArrowRightIcon, CheckIcon } from "lucide-react";
import Link from "next/link";

const alternatives = [
  {
    name: "BotGhost",
    href: "/alternatives/botghost",
    summary:
      "A no-code bot maker with many ready-made modules. Its free plan is limited to 5 servers and shows its branding in your bot's status.",
    kite: ["100 servers on the free plan", "Your own status, no branding"],
  },
  {
    name: "inventor.bot",
    href: "/alternatives/inventor",
    summary:
      "A visual bot builder that you can edit with your own AI agent. It includes 2 free bots.",
    kite: ["Up to 10 free bots", "AI built into the editor"],
  },
  {
    name: "BotWiz",
    href: "/alternatives/botwiz",
    summary:
      "A block-based bot builder that can also host Node.js and Python bots. It has a 7-day trial and then needs a subscription.",
    kite: ["A free plan with no time limit", "Up to 10 free bots"],
  },
  {
    name: "Bot Designer for Discord",
    href: "/alternatives/bdfd",
    summary:
      "A bot builder for the web and phone where you write commands in BDScript, its own scripting language.",
    kite: [
      "Blocks instead of a scripting language",
      "See every step of a command",
    ],
  },
];

export default function AlternativesPage() {
  return (
    <HomeLayout
      title="Free Alternatives to Other Discord Bot Makers"
      description="Compare Kite to BotGhost, inventor.bot, BotWiz and Bot Designer for Discord. See the free limits, features and when each one is the better choice."
    >
      <section className="container py-16 md:py-24 text-center max-w-3xl">
        <h1 className="text-4xl md:text-5xl font-bold">
          Kite Compared to Other Discord Bot Makers
        </h1>
        <p className="text-xl text-muted-foreground mt-6">
          Kite is a free, open source Discord bot maker. Here is how it compares
          to other tools for making a bot without code, including when they are
          the better choice.
        </p>
      </section>

      <section className="container pb-12 md:pb-16">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-5xl mx-auto">
          {alternatives.map((alt) => (
            <Link
              key={alt.href}
              href={alt.href}
              className="group rounded-lg border bg-muted/50 p-6 hover:border-primary transition-colors flex flex-col"
            >
              <h2 className="text-2xl font-semibold">Kite vs. {alt.name}</h2>
              <p className="text-muted-foreground mt-2">{alt.summary}</p>
              <ul className="mt-4 space-y-2 flex-auto">
                {alt.kite.map((item) => (
                  <li key={item} className="flex gap-3">
                    <CheckIcon className="flex-none w-5 h-5 mt-0.5 text-primary" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
              <div className="flex items-center gap-2 mt-6 font-medium text-primary">
                Read the comparison
                <ArrowRightIcon className="w-4 h-4 transition-transform group-hover:translate-x-1" />
              </div>
            </Link>
          ))}
        </div>
      </section>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
