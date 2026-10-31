import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import { TemplateList } from "@/components/app/TemplateList";
import { useHideAITemplates } from "@/lib/hooks/templates";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import MarketplaceBar from "@/components/home/MarketplaceBar";
import { noCodeFaq } from "@/components/landing/copy";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  categoryPath,
  templateFromBundle,
  templatePath,
  utcDay,
} from "@/lib/flow/templates";
import { Bundle } from "@/lib/types/wire.gen";
import { faqJsonLd } from "@/lib/seo";
import { fetchBundles, withApiData } from "@/lib/apiData";
import { GetStaticProps } from "next";
import { useMemo } from "react";

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
      "Open the Marketplace in your app, choose one and pick the channels and roles it should use.",
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

type Props = {
  // The official templates first, then the ones users made.
  bundles: Bundle[] | null;
  templateIconSvgs: TemplateIconSvgs;
  // When the page was rendered, see featuredTemplates.
  featuredDay: number;
};

export const getStaticProps: GetStaticProps<Props> = async () => {
  const bundles = await fetchBundles();
  // Bundles users listed show up, or leave after being taken down, within a
  // minute.
  return withApiData(
    {
      bundles,
      featuredDay: utcDay(),
      templateIconSvgs: templateIconSvgs(
        (bundles ?? []).map((b) => b.listing?.icon)
      ),
    },
    60
  );
};

export default function MarketplacePage({ bundles, featuredDay }: Props) {
  const hideAI = useHideAITemplates();
  const templates = useMemo(
    () => (bundles ?? []).flatMap((b) => templateFromBundle(b) ?? []),
    [bundles]
  );
  const count = templates.filter(
    (template) => !hideAI || !template.aiUse
  ).length;

  return (
    <HomeLayout
      title="Discord Bot Templates"
      description="Free templates for Discord bots made with Kite, from moderation commands to a verification button. Add one to your bot and change anything in the editor."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <MarketplaceBar category={null} search={false} />
      <TemplateList
        templates={templates}
        href={(t) => templatePath(t.id)}
        buildHref="/apps"
        categoryHref={categoryPath}
        featuredDay={featuredDay}
        hero={
          <div>
            <h1 className="text-3xl md:text-4xl font-bold">
              Discord Bot Templates
            </h1>
            <p className="text-lg text-muted-foreground mt-2">
              {count} templates from Kite and the community. Add one to your bot
              and change anything in the editor.
            </p>
          </div>
        }
      />

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
      <HomeCTASection light />
      <HomeFooter />
    </HomeLayout>
  );
}
