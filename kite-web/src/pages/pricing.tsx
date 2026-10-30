import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import HomePricingSection from "@/components/home/HomePricingSection";
import PlanCompareSection from "@/components/home/PlanCompareSection";
import CreditCalculator from "@/components/home/CreditCalculator";
import {
  customStatus,
  PersonaMockup,
  ProfileMockup,
  rotatingStatuses,
  ServerProfileMockup,
  StatusMockup,
} from "@/components/home/ProfileMockups";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { aiChatCompletion } from "@/lib/blocks/aiChatCompletion";
import { aiScore } from "@/lib/blocks/aiScore";
import { aiTranslate, aiTranslateStepLength } from "@/lib/blocks/aiTranslate";
import { aiWebSearch } from "@/lib/blocks/aiWebSearch";
import { controlAiCheck } from "@/lib/blocks/controlAiCheck";
import { controlConditionAiChoice } from "@/lib/blocks/controlConditionAiChoice";
import { httpRequest } from "@/lib/blocks/httpRequest";
import { BlockDefinition } from "@/lib/blocks/types";
import env from "@/lib/env/client";
import {
  aiInputStepLength,
  aiModelTiers,
  aiOutputStepTokens,
} from "@/lib/flow/aiModels";
import { CalculatorTemplate } from "@/lib/creditCalculator";
import { getCalculatorTemplates } from "@/lib/creditCalculatorTemplates";
import { faqJsonLd } from "@/lib/seo";
import { cn } from "@/lib/utils";
import { fetchBundles, fetchPlans, Plans, withApiData } from "@/lib/apiData";
import { templateFromBundle } from "@/lib/flow/templates";
import { GetStaticProps } from "next";

interface Props {
  creditCosts: { name: string; credits: string }[];
  calculatorTemplates: CalculatorTemplate[];
  plans: Plans;
}

// The credit costs come from the block definitions on the server, which keeps
// them out of the page's bundle.
export const getStaticProps: GetStaticProps<Props> = async () => {
  const [bundles, plans] = await Promise.all([fetchBundles(), fetchPlans()]);
  const official = (bundles ?? []).filter((b) => b.official);
  const tierCredits = (kind: "chat" | "search") =>
    aiModelTiers.map((t) => `${t.credits[kind]} (${t.label})`).join(", ");

  return withApiData({
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
      {
        name: `${
          aiTranslate.title
        }, for each ${aiTranslateStepLength.toLocaleString(
          "en-US"
        )} characters`,
        credits: fixedCredits(aiTranslate),
      },
      { name: aiChatCompletion.title, credits: tierCredits("chat") },
      { name: aiWebSearch.title, credits: tierCredits("search") },
      {
        name: `${aiChatCompletion.title} and ${
          aiWebSearch.title
        }, for each ${aiInputStepLength.toLocaleString(
          "en-US"
        )} bytes of prompt or ${aiOutputStepTokens} tokens of answer after the first`,
        credits: tierCredits("chat"),
      },
    ],
    calculatorTemplates: getCalculatorTemplates(
      official.flatMap((b) => templateFromBundle(b) ?? [])
    ),
    plans,
  });
};

function fixedCredits(block: BlockDefinition) {
  if (typeof block.credits !== "number") {
    throw new Error(`${block.type} has no fixed credit cost`);
  }
  return block.credits.toString();
}

const profileFeatures = [
  {
    plan: "Free",
    title: "Custom Status",
    description:
      "On every plan, your bot can show your own status and activity instead of the default one, like Watching #support.",
    mockup: <StatusMockup statuses={[customStatus]} />,
  },
  {
    plan: "Plus and up",
    title: "Rotating Status",
    description:
      "Your bot rotates through up to 10 statuses every minute, and flows can change its status with the Set status block.",
    mockup: <StatusMockup statuses={rotatingStatuses} />,
  },
  {
    plan: "Plus and up",
    title: "Animated Avatar & Banner",
    description:
      "Your bot's avatar and banner can be animated GIFs or WebPs. Change them in the settings of your bot.",
    mockup: <ProfileMockup />,
  },
  {
    plan: "Premium and up",
    title: "Profile per Server",
    description:
      "Your bot can have another nickname, avatar, banner and About Me in each server. Set them on the Servers page or from a flow with the Edit server profile block.",
    mockup: <ServerProfileMockup />,
  },
  {
    plan: "Premium and up",
    title: "Send Message As",
    description:
      "The Send message as block sends a single message with another name and avatar, like the user who ran a command or a character in a roleplay. Your bot keeps its own name and avatar everywhere else.",
    mockup: <PersonaMockup />,
  },
];

export default function PricingPage({
  creditCosts,
  calculatorTemplates,
  plans,
}: Props) {
  const faqs = [
    {
      question: "Is Kite free?",
      answer:
        "Yes. Every bot starts on the free plan, with 24/7 hosting, 10,000 credits per month and up to 100 servers. You don't need a credit card, and it isn't a trial that runs out. Premium plans raise the limits, but you don't need one to run a bot.",
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
        "The dashboard shows when your bot will run out of credits at its current rate, and warns you once it has used 80% of them. If it uses all of them, it stays online, but its commands and events pause and anyone who uses it is told why. Your bot also sends you a direct message at 80% and when it pauses. It runs again by itself when the next month begins, or a few minutes after you upgrade.",
    },
    {
      question: "Does Kite have AI features?",
      answer: `Yes. Kite's AI builder builds and changes flows from a description. The free plan includes 2 prompts per bot each month, and they don't use credits. Your bot can also use AI blocks, like Ask AI to answer questions or AI Check to answer yes or no questions, which cost credits.`,
    },
    {
      question: "Can my bot send messages with a different name and avatar?",
      answer:
        "Every bot has its own name, avatar and banner on every plan. Premium plans add the Send message as block, which sends a single message with another name and avatar, like the name and avatar of the user who ran a command or of a character in a roleplay. Your bot's own profile doesn't change. The message goes through a webhook your bot creates in the channel, so it needs the Manage Webhooks permission there.",
    },
    {
      question: "Can I cancel or get a refund?",
      answer: `You can cancel any time, and Premium keeps working until the end of the period you paid for. You can ask for a refund within 14 days of a payment by writing to ${env.NEXT_PUBLIC_CONTACT_EMAIL}.`,
    },
  ];

  return (
    <HomeLayout
      title="Pricing: Free Discord Bot Maker Plans"
      description="Kite's pricing: make a Discord bot for free with 10,000 credits per month, up to 100 servers and 15 commands per bot. No credit card. Compare the free and Premium plans and credit costs."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <HomePricingSection
        initial={plans}
        header={
          <div className="text-center">
            <h1 className="text-4xl md:text-5xl font-bold">Kite Pricing</h1>
            <p className="text-xl text-muted-foreground mt-6 max-w-3xl mx-auto">
              Kite is a free Discord bot maker. Every bot starts on the free
              plan and is hosted 24/7 with your own name, avatar and status.
              Paid plans raise the limits of a single bot when it needs more,
              and you can cancel any time.
            </p>
          </div>
        }
      />
      <ProfileFeaturesSection />

      <section className="container max-w-3xl py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          How Credits Work
        </h2>
        <p className="text-lg text-muted-foreground mt-4 text-center">
          Every time a block in a command or event listener runs, your bot uses
          the credits below. When a bot has used all of its credits, its
          commands and events pause until the next month begins. Prompts to the
          AI builder don&apos;t use credits, each plan includes a number of them
          per month.
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

      <PlanCompareSection initial={plans} />

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

function ProfileFeaturesSection() {
  return (
    <section className="container py-12 md:py-16">
      <h2 className="text-3xl md:text-4xl font-bold text-center">
        Status & Profile Features
      </h2>
      <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
        Every bot has its own name, avatar, banner and status on every plan.
        Paid plans change how it shows up in Discord beyond that.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-6 mt-10">
        {profileFeatures.map((feature, i) => (
          <Card
            key={feature.title}
            className={cn(
              "bg-muted/50 flex flex-col",
              i < 3 ? "lg:col-span-2" : "lg:col-span-3",
              i === profileFeatures.length - 1 && "md:col-span-2"
            )}
          >
            <CardHeader>
              <div className="text-sm font-medium text-primary">
                {feature.plan}
              </div>
              <CardTitle className="text-xl">{feature.title}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-1 flex-col gap-6 text-muted-foreground">
              <p>{feature.description}</p>
              <div className="mt-auto text-foreground">{feature.mockup}</div>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
