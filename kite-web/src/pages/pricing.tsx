import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import HomePricingSection from "@/components/home/HomePricingSection";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { aiChatCompletion } from "@/lib/blocks/aiChatCompletion";
import { aiScore } from "@/lib/blocks/aiScore";
import { aiWebSearch } from "@/lib/blocks/aiWebSearch";
import { controlAiCheck } from "@/lib/blocks/controlAiCheck";
import { controlConditionAiChoice } from "@/lib/blocks/controlConditionAiChoice";
import { httpRequest } from "@/lib/blocks/httpRequest";
import { BlockDefinition } from "@/lib/blocks/types";
import env from "@/lib/env/client";
import { aiModelTiers } from "@/lib/flow/aiModels";
import { faqJsonLd } from "@/lib/seo";
import { BillingPlan } from "@/lib/types/wire.gen";
import { formatInterval, formatNumber } from "@/lib/utils";
import { readFile } from "fs/promises";
import { GetStaticProps } from "next";
import Link from "next/link";
import path from "path";
import { parse } from "smol-toml";

interface Props {
  freePlan: BillingPlan;
  maxAppsPerUser: number;
  creditCosts: { name: string; credits: string }[];
}

// HomePricingSection loads the plans from the API in the browser, so crawlers
// don't see them. The free plan and the bots per user are read from the
// service's default config at build time to have them in the HTML.
export const getStaticProps: GetStaticProps<Props> = async () => {
  const file = await readFile(
    path.join(process.cwd(), "../kite-service/internal/config/default.toml"),
    "utf8"
  );
  // The TOML keys of a plan are the same as the JSON keys of the API.
  const config = parse(file) as unknown as {
    user_limits: { max_apps_per_user: number };
    billing: { plans: BillingPlan[] };
  };

  const freePlan = config.billing.plans.find((plan) => plan.default);
  if (!freePlan) throw new Error("default.toml has no default plan");

  const tierCredits = (kind: "chat" | "search") =>
    aiModelTiers.map((t) => `${t.credits[kind]} (${t.label})`).join(", ");

  return {
    props: {
      freePlan,
      maxAppsPerUser: config.user_limits.max_apps_per_user,
      creditCosts: [
        {
          name: "Most actions, like sending a message or banning a member",
          credits: "1",
        },
        { name: "Conditions and loops", credits: "Free" },
        ...[httpRequest, controlConditionAiChoice, controlAiCheck, aiScore].map(
          (block) => ({ name: block.title, credits: fixedCredits(block) })
        ),
        { name: aiChatCompletion.title, credits: tierCredits("chat") },
        { name: aiWebSearch.title, credits: tierCredits("search") },
      ],
    },
  };
};

function fixedCredits(block: BlockDefinition) {
  if (typeof block.credits !== "number") {
    throw new Error(`${block.type} has no fixed credit cost`);
  }
  return block.credits.toString();
}

export default function PricingPage({
  freePlan,
  maxAppsPerUser,
  creditCosts,
}: Props) {
  const credits = formatNumber(freePlan.feature_usage_credits_per_month);
  const servers = formatNumber(freePlan.feature_max_guilds);
  const commands = formatNumber(freePlan.feature_max_commands);
  const prompts = freePlan.feature_max_ai_prompts_per_month;

  const freePoints = [
    {
      title: `${credits} Credits per Month`,
      description:
        "Most actions cost 1 credit, and conditions and loops are free. This is enough for most servers.",
    },
    {
      title: `Up to ${servers} Servers`,
      description: `Your bot can be in up to ${servers} servers. It's hosted 24/7 and keeps your own name, avatar and status.`,
    },
    {
      title: `${commands} Commands`,
      description: `Each bot can have ${commands} slash commands, ${formatNumber(
        freePlan.feature_max_variables
      )} stored variables and ${formatNumber(
        freePlan.feature_max_messages
      )} message templates.`,
    },
    {
      title: `${freePlan.feature_max_event_listeners} Event Listeners`,
      description: `React to events like new members or messages, plus ${
        freePlan.feature_max_scheduled_event_listeners
      } schedules that run at most every ${formatInterval(
        freePlan.feature_min_schedule_interval_seconds
      )}.`,
    },
    {
      title: `${prompts} AI Builder Prompts`,
      description: `Describe a command and Kite's AI builds it. You get ${prompts} prompts per bot each month.`,
    },
    {
      title: `Up to ${maxAppsPerUser} Bots`,
      description: `You can create up to ${maxAppsPerUser} bots with one account. Every bot starts on the free plan.`,
    },
  ];

  const faqs = [
    {
      question: "Is Kite free?",
      answer: `Yes. Every bot starts on the free ${freePlan.title} plan, with 24/7 hosting, ${credits} credits per month and up to ${servers} servers. You don't need a credit card. Premium plans raise the limits, but you don't need one to run a bot.`,
    },
    {
      question: "Do I pay per bot or per account?",
      answer:
        "Per bot. A Premium subscription belongs to one bot, and the prices are per bot and month. Your other bots stay on the free plan.",
    },
    {
      question: "How many bots can I create?",
      answer: `You can create up to ${maxAppsPerUser} bots with one account. Each bot has its own plan, credits and limits.`,
    },
    {
      question: "What happens when my bot runs out of credits?",
      answer:
        "Kite stops the bot until the next month starts or until you have more credits. Then you start it again in the settings of your app.",
    },
    {
      question: "Does Kite have AI features?",
      answer: `Yes. Kite's AI builder builds and changes flows from a description. The free plan includes ${prompts} prompts per bot each month, and they don't use credits. Your bot can also use AI blocks, like Ask AI to answer questions or AI Check to answer yes or no questions, which cost credits.`,
    },
    {
      question: "Can I cancel or get a refund?",
      answer: `You can cancel any time, and Premium keeps working until the end of the period you paid for. You can ask for a refund within 14 days of a payment by writing to ${env.NEXT_PUBLIC_CONTACT_EMAIL}.`,
    },
  ];

  return (
    <HomeLayout
      title="Pricing: Free Discord Bot Maker Plans"
      description={`Kite's pricing: make a Discord bot for free with ${credits} credits per month, up to ${servers} servers and ${commands} commands per bot. No credit card. Compare the free and Premium plans and credit costs.`}
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="container max-w-5xl pt-16 md:pt-24 text-center">
        <h1 className="text-4xl md:text-5xl font-bold">Kite Pricing</h1>
        <p className="text-xl text-muted-foreground mt-6 max-w-3xl mx-auto">
          Kite is a free Discord bot maker. Every bot starts on the free{" "}
          {freePlan.title} plan and is hosted 24/7. Premium plans raise the
          limits of a single bot when it needs more.
        </p>
        <Button className="mt-8 md:w-48" asChild>
          <Link href="/apps">Get Started for Free</Link>
        </Button>
      </section>

      <HomePricingSection />

      <section className="container py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          What&apos;s Free
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
          The {freePlan.title} plan is free for every bot, with no trial that
          runs out. These are its limits per bot.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-10">
          {freePoints.map((point) => (
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

      <section className="container py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          AI Features
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-10 max-w-5xl mx-auto">
          <Card className="bg-muted/50">
            <CardHeader>
              <CardTitle className="text-xl">AI Builder</CardTitle>
            </CardHeader>
            <CardContent className="text-muted-foreground">
              Click on Ask AI in the editor and describe what a command should
              do. The AI adds and connects the blocks for you. The free plan
              includes {prompts} prompts per bot each month and Premium plans
              include more. Prompts don&apos;t use credits.
            </CardContent>
          </Card>
          <Card className="bg-muted/50">
            <CardHeader>
              <CardTitle className="text-xl">AI Blocks</CardTitle>
            </CardHeader>
            <CardContent className="text-muted-foreground">
              Your bot can use AI when it runs. Ask AI answers questions, Search
              the Web looks up current information, and AI Choice, AI Check and
              AI Score pick, check or rate something. They cost credits like
              other blocks, see the table below.
            </CardContent>
          </Card>
        </div>
      </section>

      <section className="container max-w-3xl py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          How Credits Work
        </h2>
        <p className="text-lg text-muted-foreground mt-4 text-center">
          Every time a block in a command or event listener runs, your bot uses
          the credits below. When a bot has used all of its credits, Kite stops
          it until the next month starts.
        </p>
        <table className="w-full mt-10 text-left">
          <thead>
            <tr className="border-b">
              <th className="py-3 pr-4 font-semibold">Block</th>
              <th className="py-3 font-semibold">Credits per run</th>
            </tr>
          </thead>
          <tbody>
            {creditCosts.map((cost) => (
              <tr key={cost.name} className="border-b">
                <td className="py-3 pr-4">{cost.name}</td>
                <td className="py-3 text-muted-foreground">{cost.credits}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <p className="text-muted-foreground mt-6 text-center">
          Ask AI and Search the Web cost more with smarter models.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system`}
            className="underline hover:text-foreground"
          >
            About credits
          </a>
        </p>
      </section>

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
