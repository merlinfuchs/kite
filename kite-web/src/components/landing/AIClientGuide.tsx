import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { faqJsonLd } from "@/lib/seo";
import { BadgeCheckIcon, BlocksIcon, ClockIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";
import CreateAppMockup from "./CreateAppMockup";
import GuideStep, { GuideContents } from "./GuideStep";

export interface AIClientGuideContent {
  client: string;
  title: string;
  description: string;
  heading: string;
  intro: string;
  // The steps after creating the bot and adding it to Kite.
  steps: { title: string; visual: ReactNode; body: ReactNode }[];
  prompts: string[];
  // Sections between the prompts and the FAQ, like using the client's coding
  // agent instead.
  children?: ReactNode;
  faqs: FAQProps[];
}

// Layout of the guides for building a bot from an AI client, like Claude or
// ChatGPT, through Kite's MCP server.
export default function AIClientGuide({
  content,
}: {
  content: AIClientGuideContent;
}) {
  const steps = [
    "Create Your Bot in Kite",
    ...content.steps.map((step) => step.title),
  ];

  return (
    <HomeLayout
      title={content.title}
      description={content.description}
      jsonLd={[faqJsonLd(content.faqs)]}
    >
      <section className="container max-w-5xl pt-16 md:pt-24 pb-8">
        <div className="text-sm font-semibold uppercase tracking-wide text-primary">
          <Link href="/guides" className="hover:underline">
            Guide
          </Link>
        </div>
        <h1 className="text-4xl md:text-5xl font-bold mt-3 max-w-3xl">
          {content.heading}
        </h1>
        <p className="text-xl text-muted-foreground mt-6 max-w-3xl">
          {content.intro}
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

        <GuideContents steps={steps} />
      </section>

      <div className="container max-w-5xl py-8">
        <GuideStep number={1} title={steps[0]} visual={<CreateAppMockup />}>
          <p>
            {content.client} builds your bot in Kite, and Kite hosts it. So you
            need a Discord application for the bot and a Kite app for it first.
          </p>
          <ol>
            <li>
              Create an application in the{" "}
              <a
                href="https://discord.com/developers/applications"
                target="_blank"
                rel="noreferrer noopener"
              >
                Discord Developer Portal
              </a>
              , turn on the <strong>Privileged Gateway Intents</strong> on the{" "}
              <strong>Bot</strong> page and copy the token.
            </li>
            <li>
              Open Kite, click on <strong>Create app</strong> and paste the
              token.
            </li>
            <li>
              Click on <strong>Invite app</strong> and add the bot to your
              server.
            </li>
          </ol>
          <p>
            Never done this before?{" "}
            <Link href="/guides/how-to-make-a-discord-bot#step-1">
              Our first guide
            </Link>{" "}
            shows each step in detail.
          </p>
        </GuideStep>

        {content.steps.map((step, i) => (
          <GuideStep
            key={step.title}
            number={i + 2}
            title={step.title}
            visual={step.visual}
            last={i === content.steps.length - 1}
          >
            {step.body}
          </GuideStep>
        ))}
      </div>

      <section className="container max-w-5xl py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold">
          What to Ask {content.client} For
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-3xl">
          {content.client} sees your commands, logs, channels and roles, so you
          can describe things the way you&apos;d tell a person.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-8">
          {content.prompts.map((prompt) => (
            <div
              key={prompt}
              className="rounded-lg rounded-tl-sm bg-muted px-4 py-3 font-medium"
            >
              {prompt}
            </div>
          ))}
        </div>
      </section>

      {content.children}

      <HomeFAQSection faqs={content.faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

// A section of a guide with plain text, like a short explanation.
export function GuideProse({ children }: { children: ReactNode }) {
  return (
    <section className="container py-12 md:py-16">
      <div className="prose dark:prose-invert text-foreground max-w-3xl mx-auto prose-code:before:content-none prose-code:after:content-none">
        {children}
      </div>
    </section>
  );
}
