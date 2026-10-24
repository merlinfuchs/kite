import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import HomePricingSection from "@/components/home/HomePricingSection";
import CreditCalculator from "@/components/home/CreditCalculator";
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
import { CalculatorTemplate } from "@/lib/creditCalculator";
import { getCalculatorTemplates } from "@/lib/creditCalculatorTemplates";
import { faqJsonLd } from "@/lib/seo";
import { GetStaticProps } from "next";
import Link from "next/link";

interface Props {
  creditCosts: { name: string; credits: string }[];
  calculatorTemplates: CalculatorTemplate[];
}

// HomePricingSection loads the plans from the API in the browser, so crawlers
// don't see them. The page repeats the free limits and credit costs in the
// HTML. The costs come from the block definitions at build time, which keeps
// them out of the page's bundle.
export const getStaticProps: GetStaticProps<Props> = async () => {
  const tierCredits = (kind: "chat" | "search") =>
    aiModelTiers.map((t) => `${t.credits[kind]} (${t.label})`).join(", ");

  return {
    props: {
      creditCosts: [
        {
          name: "Most actions, like sending a message or banning a member",
          credits: "1",
        },
        { name: "Conditions and loops", credits: "Free" },
        {
          name: "Getting servers, channels, members, messages and stored variables",
          credits: "Free",
        },
        ...[httpRequest, controlConditionAiChoice, controlAiCheck, aiScore].map(
          (block) => ({ name: block.title, credits: fixedCredits(block) })
        ),
        { name: aiChatCompletion.title, credits: tierCredits("chat") },
        { name: aiWebSearch.title, credits: tierCredits("search") },
      ],
      calculatorTemplates: getCalculatorTemplates(),
    },
  };
};

function fixedCredits(block: BlockDefinition) {
  if (typeof block.credits !== "number") {
    throw new Error(`${block.type} has no fixed credit cost`);
  }
  return block.credits.toString();
}

const freePoints = [
  {
    title: "10,000 Credits per Month",
    description:
      "Most actions cost 1 credit. Conditions, loops and reading data like members or messages are free. This is enough for most servers.",
  },
  {
    title: "Up to 100 Servers",
    description:
      "Your bot can be in up to 100 servers. It's hosted 24/7 and keeps your own name, avatar and status.",
  },
  {
    title: "25 Commands",
    description:
      "Each bot can have 25 slash commands, 25 stored variables and 25 message templates.",
  },
  {
    title: "5 Event Listeners",
    description:
      "React to events like new members or messages, plus 5 schedules that run at most every 5 minutes.",
  },
  {
    title: "2 AI Builder Prompts",
    description:
      "Describe a command and Kite's AI builds it. You get 2 prompts per bot each month.",
  },
  {
    title: "Up to 10 Bots",
    description:
      "You can create up to 10 bots with one account. Every bot starts on the free plan.",
  },
];

export default function PricingPage({
  creditCosts,
  calculatorTemplates,
}: Props) {
  const faqs = [
    {
      question: "Is Kite free?",
      answer:
        "Yes. Every bot starts on the free plan, with 24/7 hosting, 10,000 credits per month and up to 100 servers. You don't need a credit card. Premium plans raise the limits, but you don't need one to run a bot.",
    },
    {
      question: "Do I pay per bot or per account?",
      answer:
        "Per bot. A Premium subscription belongs to one bot, and the prices are per bot and month. Your other bots stay on the free plan.",
    },
    {
      question: "How many bots can I create?",
      answer:
        "You can create up to 10 bots with one account. Each bot has its own plan, credits and limits.",
    },
    {
      question: "What happens when my bot runs out of credits?",
      answer:
        "The dashboard shows when your bot will run out of credits at its current rate, and warns you once it has used 80% of them. If it uses all of them, Kite stops it and starts it again by itself when the next month begins, or a few minutes after you upgrade.",
    },
    {
      question: "Does Kite have AI features?",
      answer: `Yes. Kite's AI builder builds and changes flows from a description. The free plan includes 2 prompts per bot each month, and they don't use credits. Your bot can also use AI blocks, like Ask AI to answer questions or AI Check to answer yes or no questions, which cost credits.`,
    },
    {
      question: "Can I cancel or get a refund?",
      answer: `You can cancel any time, and Premium keeps working until the end of the period you paid for. You can ask for a refund within 14 days of a payment by writing to ${env.NEXT_PUBLIC_CONTACT_EMAIL}.`,
    },
  ];

  return (
    <HomeLayout
      title="Pricing: Free Discord Bot Maker Plans"
      description="Kite's pricing: make a Discord bot for free with 10,000 credits per month, up to 100 servers and 25 commands per bot. No credit card. Compare the free and Premium plans and credit costs."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="container max-w-5xl pt-16 md:pt-24 text-center">
        <h1 className="text-4xl md:text-5xl font-bold">Kite Pricing</h1>
        <p className="text-xl text-muted-foreground mt-6 max-w-3xl mx-auto">
          Kite is a free Discord bot maker. Every bot starts on the free plan
          and is hosted 24/7. Premium plans raise the limits of a single bot
          when it needs more.
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
          Every bot gets the free plan, with no trial that runs out. These are
          its limits per bot.
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
              includes 2 prompts per bot each month, and Premium plans include
              more. Prompts don&apos;t use credits.
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
          it and starts it again by itself when the next month begins.
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

      <section className="container max-w-3xl py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          Estimate Your Usage
        </h2>
        <p className="text-lg text-muted-foreground mt-4 text-center">
          Pick what your bot will do and how big your server is to see if the
          free plan is enough. The costs come from Kite&apos;s templates.
        </p>
        <CreditCalculator templates={calculatorTemplates} />
      </section>

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
