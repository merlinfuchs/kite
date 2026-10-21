import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import CreateAppMockup from "@/components/landing/CreateAppMockup";
import DevPortalMockup from "@/components/landing/DevPortalMockup";
import DiscordReplyMockup from "@/components/landing/DiscordReplyMockup";
import { helloFlow } from "@/components/landing/flows";
import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import { faqJsonLd } from "@/lib/seo";
import dynamic from "next/dynamic";
import { BadgeCheckIcon, BlocksIcon, ClockIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";

const FlowExample = dynamic(() => import("@/components/flow/FlowExample"), {
  ssr: false,
});

const steps = [
  "Create a Discord Application",
  "Add Your Bot to Kite",
  "Build Your First Command",
  "Try It in Discord",
];

const faqs = [
  {
    question: "Can I make a Discord bot for free?",
    answer:
      "Yes. Creating a Discord application is free, and Kite builds and hosts your bot for free. The free plan includes 10,000 credits per month and up to 100 servers.",
  },
  {
    question: "Do I need to know how to code to make a Discord bot?",
    answer:
      "No. With a bot maker like Kite you build commands by connecting blocks. You only need code if you want to write and host the bot yourself.",
  },
  {
    question: "How long does it take to make a Discord bot?",
    answer:
      "Creating the application and your first command takes about 10 minutes. New commands can take up to a minute to show up in Discord.",
  },
  {
    question: "Can I make a Discord bot on my phone?",
    answer:
      "You can create the application and add it to Kite on your phone. Kite's editor works best on a computer, because you drag and connect blocks.",
  },
];

export default function HowToMakeADiscordBotPage() {
  return (
    <HomeLayout
      title="How to Make a Discord Bot Without Coding"
      description="Learn how to make your own Discord bot for free without coding. Create a Discord application, add it to Kite and build your first slash command in about 10 minutes."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="container max-w-5xl pt-16 md:pt-24 pb-8">
        <div className="text-sm font-semibold uppercase tracking-wide text-primary">
          Guide
        </div>
        <h1 className="text-4xl md:text-5xl font-bold mt-3 max-w-3xl">
          How to Make a Discord Bot Without Coding
        </h1>
        <p className="text-xl text-muted-foreground mt-6 max-w-3xl">
          You can make your own Discord bot for free without writing any code.
          This guide shows you how to create the bot in Discord, add it to Kite
          and build your first slash command.
        </p>
        <div className="flex flex-wrap gap-2 mt-6">
          {[
            { icon: ClockIcon, label: "About 10 minutes" },
            { icon: BadgeCheckIcon, label: "Free" },
            { icon: BlocksIcon, label: "No code" },
          ].map((tag) => (
            <div
              key={tag.label}
              className="flex items-center gap-2 rounded-full border px-3 py-1 text-sm text-muted-foreground"
            >
              <tag.icon className="w-4 h-4" />
              {tag.label}
            </div>
          ))}
        </div>
        <Button className="mt-8 md:w-48" asChild>
          <Link href="/apps">Get Started for Free</Link>
        </Button>

        <nav className="mt-12 rounded-lg border bg-muted/50 p-6">
          <div className="font-semibold">In this guide</div>
          <ol className="mt-3 grid sm:grid-cols-2 gap-2">
            {steps.map((title, i) => (
              <li key={title}>
                <a
                  href={`#step-${i + 1}`}
                  className="text-muted-foreground hover:text-foreground"
                >
                  {i + 1}. {title}
                </a>
              </li>
            ))}
          </ol>
        </nav>
      </section>

      <div className="container max-w-5xl py-8">
        <Step number={1} title={steps[0]} visual={<DevPortalMockup />}>
          <p>
            Every Discord bot belongs to an application. The application holds
            the name, avatar and token of your bot.
          </p>
          <ol>
            <li>
              Open the{" "}
              <a
                href="https://discord.com/developers/applications"
                target="_blank"
                rel="noreferrer noopener"
              >
                Discord Developer Portal
              </a>{" "}
              and sign in.
            </li>
            <li>
              Click on <strong>New Application</strong>, give your bot a name
              and upload an avatar.
            </li>
            <li>
              Open the <strong>Bot</strong> page and turn on the three{" "}
              <strong>Privileged Gateway Intents</strong>. Without them your bot
              can&apos;t see new members or read messages.
            </li>
            <li>
              Click on <strong>Reset Token</strong> and copy the token.
            </li>
          </ol>
          <p>
            Keep the token secret. Anyone who has it can control your bot. If
            you share it by accident, reset it and use the new one.
          </p>
        </Step>

        <Step number={2} title={steps[1]} visual={<CreateAppMockup />}>
          <p>
            Kite hosts your bot and gives you an editor to build its commands.
          </p>
          <ol>
            <li>Open Kite and sign in with Discord.</li>
            <li>
              Click on <strong>Create app</strong>, paste your token and
              confirm.
            </li>
            <li>
              Click on <strong>Invite app</strong> at the top right and choose
              your server.
            </li>
          </ol>
          <p>
            Your bot now shows up in your server. It&apos;s online, but it
            doesn&apos;t do anything yet.
          </p>
          <Button asChild className="not-prose">
            <Link href="/apps">Open Kite</Link>
          </Button>
        </Step>

        <Step
          number={3}
          title={steps[2]}
          visual={
            <div className="h-[340px] rounded-lg border bg-muted/50">
              <FlowExample nodes={helloFlow.nodes} edges={helloFlow.edges} />
            </div>
          }
        >
          <p>
            A slash command is what members type to use your bot, like{" "}
            <code>/hello</code>. In Kite you build each command by connecting
            blocks.
          </p>
          <ol>
            <li>
              Open <strong>Commands</strong> in the sidebar and click on{" "}
              <strong>Create command</strong>. Name it <code>hello</code>.
            </li>
            <li>
              Drag a <strong>Create response message</strong> block into the
              editor and connect it to the command block.
            </li>
            <li>
              Click on the new block and enter a message, like{" "}
              <code>Hello {"{{user.mention}}"}!</code>
            </li>
            <li>
              Click on <strong>Save Changes</strong>. Kite updates the command
              in Discord a few seconds later.
            </li>
          </ol>
          <p>
            Stuck? Click on <strong>Ask AI</strong> in the editor and describe
            what the command should do.
          </p>
        </Step>

        <Step
          number={4}
          last
          title={steps[3]}
          visual={<DiscordReplyMockup command="hello" content="Hello @Alex!" />}
        >
          <p>
            Type <code>/hello</code> in your server. Your bot answers right
            away.
          </p>
          <p>
            A new command can take up to a minute to appear. Restart Discord if
            you don&apos;t see it.
          </p>
        </Step>
      </div>

      <section className="container py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          What to Build Next
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
          Templates add finished commands to your app in a few clicks. You can
          change them afterwards.
        </p>
        <div className="grid md:grid-cols-3 gap-6 mt-10">
          {nextSteps.map((next) => (
            <Link
              key={next.href}
              href={next.href}
              className="rounded-lg border bg-muted/50 p-6 hover:border-primary transition-colors"
            >
              <div className="text-xl font-semibold">{next.title}</div>
              <div className="text-muted-foreground mt-2">
                {next.description}
              </div>
            </Link>
          ))}
        </div>
      </section>

      <section className="container py-12 md:py-16">
        <div className="prose dark:prose-invert text-foreground max-w-3xl mx-auto">
          <h2>Making a Discord Bot With Code</h2>
          <p>
            You can also write a bot yourself with a library like discord.js for
            JavaScript or discord.py for Python. This gives you full control,
            but you need to know the language, handle errors and keep a server
            running so the bot stays online. For most community servers, a bot
            maker is faster to set up and easier to change.
          </p>
          <p>
            For more details on Kite, read the{" "}
            <a href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/getting-started`}>
              getting started guide
            </a>{" "}
            in the documentation.
          </p>
        </div>
      </section>

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

const nextSteps = [
  {
    title: "Moderation Bot",
    description: "Add /ban, /kick and /mute commands from a template.",
    href: "/discord-moderation-bot",
  },
  {
    title: "Welcome Bot",
    description: "Greet new members and give them a role.",
    href: "/discord-welcome-bot",
  },
  {
    title: "AI Bot",
    description: "Answer questions with an /ask command.",
    href: "/ai-discord-bot",
  },
];

function Step({
  number,
  title,
  visual,
  last,
  children,
}: {
  number: number;
  title: string;
  visual: ReactNode;
  last?: boolean;
  children: ReactNode;
}) {
  return (
    <section
      id={`step-${number}`}
      className="relative pl-14 pb-16 scroll-mt-20"
    >
      <div className="absolute left-0 top-0 w-10 h-10 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
        {number}
      </div>
      {!last && (
        <div className="absolute left-5 top-12 bottom-2 border-l-2 border-dashed" />
      )}
      <h2 className="text-2xl md:text-3xl font-bold leading-10">{title}</h2>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-10 mt-4">
        <div className="prose dark:prose-invert text-foreground prose-code:before:content-none prose-code:after:content-none">
          {children}
        </div>
        <div>{visual}</div>
      </div>
    </section>
  );
}
