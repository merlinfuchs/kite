import {
  ArrowRightIcon,
  ServerIcon,
  ShieldIcon,
  SparklesIcon,
  UserPlusIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import Link from "next/link";

const useCases = [
  {
    title: "Moderation Bot",
    description: "Ban, kick and mute members with commands that you control.",
    icon: ShieldIcon,
    color: "bg-red-500/15 text-red-600 dark:text-red-400",
    href: "/discord-moderation-bot",
  },
  {
    title: "Welcome Bot",
    description:
      "Greet new members, give them a role and send them your rules.",
    icon: UserPlusIcon,
    color: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400",
    href: "/discord-welcome-bot",
  },
  {
    title: "AI Chatbot",
    description:
      "Answer questions with an /ask command and its own personality.",
    icon: SparklesIcon,
    color: "bg-violet-500/15 text-violet-600 dark:text-violet-400",
    href: "/ai-discord-bot",
  },
  {
    title: "Anything Else",
    description: "Build your own commands and events, hosted 24/7 for free.",
    icon: ServerIcon,
    color: "bg-blue-500/15 text-blue-600 dark:text-blue-400",
    href: "/discord-bot-hosting",
  },
];

export default function HomeUseCasesSection() {
  return (
    <section id="use-cases" className="container py-16 sm:py-24">
      <h2 className="text-3xl md:text-4xl font-bold">
        What Will You{" "}
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Build?
        </span>
      </h2>
      <p className="text-lg md:text-xl text-muted-foreground mt-4 max-w-2xl">
        Start from a template for the most common bots, or build something
        nobody has built before.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4 mt-8 md:mt-10">
        {useCases.map((useCase) => (
          <Link
            key={useCase.href}
            href={useCase.href}
            className="group rounded-xl border bg-card p-4 md:p-6 transition-all hover:border-primary/50 hover:shadow-md flex md:flex-col gap-4 md:gap-0"
          >
            <div
              className={`flex-none w-12 h-12 rounded-lg flex items-center justify-center ${useCase.color}`}
            >
              <useCase.icon className="w-6 h-6" />
            </div>
            <div className="flex-auto flex flex-col">
              <h3 className="text-lg md:text-xl font-semibold md:mt-4">
                {useCase.title}
              </h3>
              <p className="text-muted-foreground mt-1 md:mt-2 flex-auto">
                {useCase.description}
              </p>
              {/* On phones the whole card is the link, so this only shows from md up. */}
              <div className="hidden md:flex items-center gap-2 mt-4 text-sm font-medium text-primary">
                Learn more
                <ArrowRightIcon className="w-4 h-4 transition-transform group-hover:translate-x-1" />
              </div>
            </div>
          </Link>
        ))}
      </div>

      <div className="mt-6 rounded-xl border bg-muted/40 p-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="font-semibold">Coming from another bot maker?</div>
          <div className="text-muted-foreground">
            See how Kite compares to BotGhost, inventor.bot, BotWiz and Bot
            Designer for Discord.
          </div>
        </div>
        <Button asChild variant="outline" className="flex-none">
          <Link href="/alternatives">Compare Bot Makers</Link>
        </Button>
      </div>
    </section>
  );
}
