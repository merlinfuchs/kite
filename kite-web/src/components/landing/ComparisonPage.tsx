import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { faqJsonLd } from "@/lib/seo";
import { CheckIcon } from "lucide-react";
import Link from "next/link";
import { banFlow } from "./flows";
import ShowcaseVisual from "./ShowcaseVisual";

export interface ComparisonPageContent {
  title: string;
  description: string;
  competitor: string;
  heading: string;
  intro: string;
  chooseKite: string[];
  chooseCompetitor: string[];
  rows: { feature: string; kite: string; competitor: string }[];
  reasons: { title: string; description: string }[];
  faqs: FAQProps[];
}

// Layout of the pages that compare Kite to another bot maker.
export default function ComparisonPage({
  content,
}: {
  content: ComparisonPageContent;
}) {
  const { competitor } = content;

  return (
    <HomeLayout
      title={content.title}
      description={content.description}
      jsonLd={[faqJsonLd(content.faqs)]}
    >
      <section className="container py-16 md:py-24">
        <div className="text-center max-w-3xl mx-auto">
          <div className="text-sm font-semibold uppercase tracking-wide text-primary">
            Kite vs. {competitor}
          </div>
          <h1 className="text-4xl md:text-5xl font-bold mt-3">
            {content.heading}
          </h1>
          <p className="text-xl text-muted-foreground mt-6">{content.intro}</p>
          <Button className="mt-8 md:w-48" asChild>
            <Link href="/apps">Get Started for Free</Link>
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-14 max-w-4xl mx-auto">
          <Verdict
            title="Choose Kite if"
            items={content.chooseKite}
            highlight
          />
          <Verdict
            title={`Choose ${competitor} if`}
            items={content.chooseCompetitor}
          />
        </div>
      </section>

      <section className="container py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center mb-10">
          Kite and {competitor} Compared
        </h2>
        <div className="max-w-4xl mx-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead></TableHead>
                <TableHead>Kite</TableHead>
                <TableHead>{competitor}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {content.rows.map((row) => (
                <TableRow key={row.feature}>
                  <TableCell className="font-medium">{row.feature}</TableCell>
                  <TableCell>{row.kite}</TableCell>
                  <TableCell>{row.competitor}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </section>

      <section className="container py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          Why Use Kite Instead of {competitor}
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-8 mt-10 max-w-5xl mx-auto">
          {content.reasons.map((reason) => (
            <div key={reason.title} className="flex gap-4">
              <CheckIcon className="flex-none w-6 h-6 text-primary mt-1" />
              <div>
                <h3 className="text-xl font-semibold">{reason.title}</h3>
                <p className="text-muted-foreground mt-1">
                  {reason.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="container grid grid-cols-1 lg:grid-cols-2 gap-12 items-center py-12 md:py-16">
        <div>
          <h2 className="text-3xl md:text-4xl font-bold">
            Build Commands by Connecting Blocks
          </h2>
          <p className="text-lg text-muted-foreground mt-4">
            This is the /ban command from the moderation template. Each block is
            one step, and the lines between them set the order. You can import
            it with one click and change it however you like.
          </p>
        </div>
        <ShowcaseVisual
          showcase={{
            flow: banFlow,
            command: "ban",
            reply: "Banned the user.",
          }}
        />
      </section>

      <section className="container py-12 md:py-16">
        <h2 className="text-3xl md:text-4xl font-bold text-center">
          How to Switch From {competitor}
        </h2>
        <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
          You keep your Discord application, so your bot keeps its name, avatar
          and servers.
        </p>
        <ol className="grid md:grid-cols-3 gap-8 mt-10 max-w-5xl mx-auto">
          {[
            {
              title: "Stop Your Bot",
              description: `Turn your bot off in ${competitor}, so the two don't answer at the same time.`,
            },
            {
              title: "Add It to Kite",
              description:
                "Copy the bot token from the Discord Developer Portal, sign in to Kite and click on Create app.",
            },
            {
              title: "Build Your Commands",
              description:
                "Import a template, build commands from blocks or describe them to Kite's AI. They show up in Discord when you save them.",
            },
          ].map((step, i) => (
            <li key={step.title} className="text-center">
              <div className="w-10 h-10 mx-auto rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                {i + 1}
              </div>
              <h3 className="text-xl font-semibold mt-4">{step.title}</h3>
              <p className="text-muted-foreground mt-2">{step.description}</p>
            </li>
          ))}
        </ol>
      </section>

      <HomeFAQSection faqs={content.faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

function Verdict({
  title,
  items,
  highlight,
}: {
  title: string;
  items: string[];
  highlight?: boolean;
}) {
  return (
    <div
      className={
        highlight
          ? "rounded-lg border-2 border-primary bg-primary/5 p-6"
          : "rounded-lg border bg-muted/50 p-6"
      }
    >
      <h2 className="text-xl font-semibold">{title}</h2>
      <ul className="mt-4 space-y-3">
        {items.map((item) => (
          <li key={item} className="flex gap-3">
            <CheckIcon
              className={
                highlight
                  ? "flex-none w-5 h-5 mt-0.5 text-primary"
                  : "flex-none w-5 h-5 mt-0.5 text-muted-foreground"
              }
            />
            <span>{item}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
