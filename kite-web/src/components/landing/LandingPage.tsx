import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { faqJsonLd } from "@/lib/seo";
import Link from "next/link";
import { Fragment } from "react";
import Showcase, { ShowcaseContent } from "./Showcase";

export interface LandingPageContent {
  title: string;
  description: string;
  heading: string;
  intro: string;
  secondaryAction?: { label: string; href: string };
  showcase?: ShowcaseContent;
  sections: {
    heading: string;
    body?: string;
    points?: { title: string; description: string }[];
  }[];
  comparison?: {
    competitor: string;
    rows: { feature: string; kite: string; competitor: string }[];
  };
  faqs: FAQProps[];
}

// Copy that several landing pages share.
export const openSourcePoint = {
  title: "Open Source",
  description:
    "Kite's code is on GitHub. You can see how your bot runs, or host Kite yourself.",
};

export const aiBuilderPoint = {
  title: "Let AI Build It",
  description:
    "Describe what a command should do and Kite's AI adds and connects the blocks for you.",
};

export const freeFaq: FAQProps = {
  question: "Is Kite really free?",
  answer:
    "Yes. Your bot is hosted 24/7 for free. Premium plans give you higher limits, like more servers and commands, but you don't need one to run a bot.",
};

export const noCodeFaq: FAQProps = {
  question: "Do I need to know how to code?",
  answer:
    "No. You build commands by connecting blocks in a visual editor. You can also describe a command and Kite's AI builds it for you.",
};

export function switchSection(competitor: string) {
  return {
    heading: `How to Switch From ${competitor}`,
    body: "You keep your Discord application, so your bot keeps its name, avatar and servers.",
    points: [
      {
        title: "1. Stop Your Bot",
        description: `Turn your bot off in ${competitor}, so the two don't answer at the same time.`,
      },
      {
        title: "2. Add It to Kite",
        description:
          "Copy the bot token from the Discord Developer Portal, sign in to Kite and click on Create app.",
      },
      {
        title: "3. Build Your Commands",
        description:
          "Import a template, build commands from blocks or describe them to Kite's AI. Then deploy them.",
      },
    ],
  };
}

export default function LandingPage({
  content,
}: {
  content: LandingPageContent;
}) {
  const comparison = content.comparison && (
    <section className="container py-12 md:py-16">
      <h2 className="text-3xl md:text-4xl font-bold text-center mb-10">
        Kite vs. {content.comparison.competitor}
      </h2>
      <div className="max-w-4xl mx-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead></TableHead>
              <TableHead>Kite</TableHead>
              <TableHead>{content.comparison.competitor}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {content.comparison.rows.map((row) => (
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
  );

  return (
    <HomeLayout
      title={content.title}
      description={content.description}
      jsonLd={[faqJsonLd(content.faqs)]}
    >
      <section className="container py-20 md:py-28 text-center">
        <h1 className="text-4xl md:text-5xl font-bold max-w-4xl mx-auto">
          {content.heading}
        </h1>
        <p className="text-xl text-muted-foreground mt-6 max-w-3xl mx-auto">
          {content.intro}
        </p>
        <div className="mt-8 flex flex-col md:flex-row justify-center gap-4">
          <Button className="md:w-48" asChild>
            <Link href="/apps">Get Started for Free</Link>
          </Button>
          {content.secondaryAction && (
            <Button variant="outline" className="md:w-48" asChild>
              <a href={content.secondaryAction.href}>
                {content.secondaryAction.label}
              </a>
            </Button>
          )}
        </div>
      </section>

      {content.showcase && <Showcase content={content.showcase} />}

      {content.sections.map((section, i) => (
        <Fragment key={section.heading}>
          <section className="container py-12 md:py-16">
            <h2 className="text-3xl md:text-4xl font-bold text-center">
              {section.heading}
            </h2>
            {section.body && (
              <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
                {section.body}
              </p>
            )}
            {section.points && (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-10">
                {section.points.map((point) => (
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
            )}
          </section>
          {/* The table backs up the reasons in the first section. */}
          {i === 0 && comparison}
        </Fragment>
      ))}

      <HomeFAQSection faqs={content.faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
