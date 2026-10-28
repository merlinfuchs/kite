import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import env from "@/lib/env/client";
import {
  ArrowRightIcon,
  BlocksIcon,
  BotIcon,
  SparklesIcon,
} from "lucide-react";
import Link from "next/link";

const guides = [
  {
    icon: BlocksIcon,
    title: "How to Make a Discord Bot Without Coding",
    description:
      "Create a Discord application, add it to Kite and build your first slash command in the editor.",
    href: "/guides/how-to-make-a-discord-bot",
  },
  {
    icon: SparklesIcon,
    title: "How to Make a Discord Bot With Claude",
    description:
      "Connect Claude to Kite and let it build the commands and events of your bot.",
    href: "/guides/make-a-discord-bot-with-claude",
  },
  {
    icon: BotIcon,
    title: "How to Make a Discord Bot With ChatGPT",
    description:
      "Connect ChatGPT to Kite and describe your bot instead of copying code.",
    href: "/guides/make-a-discord-bot-with-chatgpt",
  },
];

export default function GuidesPage() {
  return (
    <HomeLayout
      title="Discord Bot Guides"
      description="Step-by-step guides for making your own Discord bot for free: without coding, with Claude or with ChatGPT."
    >
      <section className="container max-w-5xl pt-16 md:pt-24 pb-12">
        <h1 className="text-4xl md:text-5xl font-bold">Discord Bot Guides</h1>
        <p className="text-xl text-muted-foreground mt-6 max-w-3xl">
          Make your own Discord bot for free, step by step. Build it in
          Kite&apos;s editor, or let Claude or ChatGPT build it for you.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-12">
          {guides.map((guide) => (
            <Link
              key={guide.href}
              href={guide.href}
              className="group flex flex-col rounded-lg border bg-muted/50 p-6 hover:border-primary transition-colors"
            >
              <div className="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
                <guide.icon className="size-5" />
              </div>
              <div className="text-xl font-semibold mt-4">{guide.title}</div>
              <div className="text-muted-foreground mt-2 flex-auto">
                {guide.description}
              </div>
              <div className="flex items-center gap-1 text-primary font-medium mt-4">
                Read the guide
                <ArrowRightIcon className="size-4 transition-transform group-hover:translate-x-0.5" />
              </div>
            </Link>
          ))}
        </div>
        <p className="text-muted-foreground mt-10">
          Looking for how a block or feature works? The{" "}
          <a
            href={env.NEXT_PUBLIC_DOCS_LINK}
            className="text-primary hover:underline"
          >
            documentation
          </a>{" "}
          covers everything in Kite.
        </p>
      </section>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
