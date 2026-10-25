import { TemplateList } from "@/components/app/TemplateList";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { noCodeFaq } from "@/components/landing/copy";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getTemplates } from "@/lib/flow/templateFiles";
import { faqJsonLd } from "@/lib/seo";
import { SparklesIcon } from "lucide-react";
import Link from "next/link";

const points = [
  {
    title: "Change Every Block",
    description:
      "A template is copied into your bot as normal commands and event listeners. Open them in the editor and change the replies, permissions or anything else.",
  },
  {
    title: "Add Only What You Need",
    description:
      "Pick the commands and listeners you want from a template. Commands your bot already has are left alone.",
  },
  {
    title: "Combine Templates",
    description:
      "Add several templates to the same bot, like Moderation, Welcomer and Verification, and build your own commands next to them.",
  },
  {
    title: "Pick Channels and Roles",
    description:
      "Choose the channels and roles a template uses from your server when you add it, instead of copying IDs.",
  },
];

const steps = [
  {
    title: "Create Your Bot",
    description:
      "Sign in with Discord, add your bot's token to Kite and invite the bot to your server.",
  },
  {
    title: "Add a Template",
    description:
      "Open Templates in your app, choose one and pick the channels and roles it should use.",
  },
  {
    title: "Make It Your Own",
    description:
      "The commands show up in Discord a few seconds later. Change them in the editor whenever you like.",
  },
];

const faqs: FAQProps[] = [
  {
    question: "What is a Discord bot template?",
    answer:
      "A ready-made set of slash commands, event listeners and messages that you add to your bot in Kite. Once added, they're yours: every block can be changed in the editor like anything you built yourself.",
  },
  {
    question: "Are the templates free?",
    answer:
      "Yes. Every template works on the free plan, which includes 24/7 hosting and 10,000 credits per month for everything your bot does.",
  },
  noCodeFaq,
  {
    question: "Will my changes be kept?",
    answer:
      "Yes. A template is copied into your bot when you add it, so the commands are yours to change and nothing overwrites them.",
  },
  {
    question: "Can I add more than one template to my bot?",
    answer:
      "Yes. Add as many as you like to the same bot. When you add one again, Kite only adds the parts your bot doesn't have yet.",
  },
  {
    question: "What if there's no template for what I need?",
    answer:
      "Describe the command you want and Kite's AI builds it for you, or connect the blocks yourself in the editor.",
  },
];

export default function TemplatesPage() {
  const count = getTemplates().length;

  return (
    <HomeLayout
      title="Discord Bot Templates"
      description="Free templates for Discord bots made with Kite, from moderation commands to a verification button. Add one to your bot and change anything in the editor."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="border-b bg-muted/30">
        <div className="container max-w-5xl py-14 md:py-20 text-center">
          <h1 className="text-4xl md:text-5xl font-bold">
            Discord Bot Templates
          </h1>
          <p className="text-xl text-muted-foreground mt-4 max-w-2xl mx-auto">
            Ready-made commands, event listeners and messages for your bot. Add
            one in Kite and change it like anything you built yourself.
          </p>
          <div className="mt-8 flex flex-col sm:flex-row justify-center gap-4">
            <Button className="sm:w-48" asChild>
              <Link href="/apps">Get Started for Free</Link>
            </Button>
            <Button variant="outline" className="sm:w-48" asChild>
              <a href="#templates">Browse {count} Templates</a>
            </Button>
          </div>
        </div>
      </section>

      <section
        id="templates"
        className="container max-w-5xl py-12 scroll-mt-16"
      >
        <TemplateList
          templates={getTemplates()}
          href={(t) => `/templates/${t.id}`}
          last={<BuildYourOwnCard />}
        />
      </section>

      <section className="container max-w-5xl py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          What You Can Do With Templates
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
          Templates save you building the common parts of a bot, so you can
          spend your time on what makes your server different.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-10">
          {points.map((point) => (
            <Card key={point.title} className="bg-muted/50">
              <CardHeader>
                <CardTitle className="text-xl">{point.title}</CardTitle>
              </CardHeader>
              <CardContent className="text-muted-foreground">
                {point.description}
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <section className="container max-w-5xl py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          How to Add a Template
        </h2>
        <ol className="relative grid md:grid-cols-3 gap-8 mt-12">
          {/* Connects the numbers of the steps. */}
          <div className="hidden md:block absolute top-6 left-[16.7%] right-[16.7%] border-t-2 border-dashed" />
          {steps.map((step, i) => (
            <li key={step.title} className="text-center">
              <div className="relative w-12 h-12 mx-auto rounded-full bg-primary text-primary-foreground flex items-center justify-center text-lg font-bold">
                {i + 1}
              </div>
              <h3 className="text-xl font-semibold mt-4">{step.title}</h3>
              <p className="text-muted-foreground mt-2">{step.description}</p>
            </li>
          ))}
        </ol>
      </section>

      <HomeFAQSection faqs={faqs} className="max-w-5xl" />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

function BuildYourOwnCard() {
  return (
    <Link href="/apps" className="group">
      <Card className="h-full flex flex-col p-5 border-dashed bg-transparent transition-colors group-hover:border-primary/60 group-hover:bg-muted/30">
        <div className="h-11 w-11 bg-primary/15 flex-none rounded-lg flex items-center justify-center mb-4">
          <SparklesIcon className="w-6 h-6 text-primary" />
        </div>
        <div className="font-semibold mb-1">Build Your Own</div>
        <div className="text-sm text-muted-foreground">
          Don&apos;t see what you need? Describe a command and Kite&apos;s AI
          builds it, or connect the blocks yourself.
        </div>
      </Card>
    </Link>
  );
}
