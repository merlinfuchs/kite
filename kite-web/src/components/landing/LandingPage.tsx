import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { faqJsonLd } from "@/lib/seo";
import { cn } from "@/lib/utils";
import { ArrowRightIcon } from "lucide-react";
import Link from "next/link";
import ShowcaseVisual, { Showcase } from "./ShowcaseVisual";

export interface LandingPageContent {
  title: string;
  description: string;
  heading: string;
  intro: string;
  secondaryAction?: { label: string; href: string };
  showcase: Showcase;
  sections: {
    heading: string;
    body?: string;
    // Points with an href link to a page that tells more, like a guide.
    points?: { title: string; description: string; href?: string }[];
  }[];
  faqs: FAQProps[];
}

// Titles of steps start with their number, like "1. Create Your Bot".
const stepTitle = /^\d+\.\s+/;

function CommandList({
  points,
}: {
  points: { title: string; description: string }[];
}) {
  return (
    <div className="rounded-xl border bg-card p-2 shadow-sm">
      <div className="px-3 pt-2 pb-1 text-xs font-semibold uppercase text-muted-foreground">
        Your Bot
      </div>
      {points.map((point, i) => (
        <div
          key={point.title}
          className={cn("rounded-md px-3 py-2", i === 0 && "bg-muted")}
        >
          <div className="font-semibold">{point.title}</div>
          <div className="text-sm text-muted-foreground">
            {point.description}
          </div>
        </div>
      ))}
    </div>
  );
}

// Layout of the pages for a use case, like a moderation bot or bot hosting.
export default function LandingPage({
  content,
}: {
  content: LandingPageContent;
}) {
  return (
    <HomeLayout
      title={content.title}
      description={content.description}
      jsonLd={[faqJsonLd(content.faqs)]}
    >
      <section className="container grid grid-cols-1 lg:grid-cols-2 gap-12 items-center py-16 md:py-24">
        <div className="text-center lg:text-start">
          <h1 className="text-4xl md:text-5xl font-bold">{content.heading}</h1>
          <p className="text-xl text-muted-foreground mt-6">{content.intro}</p>
          <div className="mt-8 flex flex-col md:flex-row justify-center lg:justify-start gap-4">
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
        </div>
        <ShowcaseVisual showcase={content.showcase} />
      </section>

      {content.sections.map((section) => {
        const points = section.points ?? [];
        const header = (
          <>
            <h2 className="text-3xl md:text-4xl font-bold">
              {section.heading}
            </h2>
            {section.body && (
              <p className="text-lg text-muted-foreground mt-4 max-w-2xl">
                {section.body}
              </p>
            )}
          </>
        );

        // The commands a template adds, shown like Discord lists them.
        if (points.length > 0 && points.every((p) => p.title.startsWith("/"))) {
          return (
            <section
              key={section.heading}
              className="container py-12 md:py-16 grid grid-cols-1 lg:grid-cols-2 gap-10 items-center"
            >
              <div>{header}</div>
              <CommandList points={points} />
            </section>
          );
        }

        const steps = points.every((p) => stepTitle.test(p.title));
        const List = steps ? "ol" : "ul";
        return (
          <section key={section.heading} className="container py-12 md:py-16">
            {header}
            {points.length > 0 && (
              <List
                className={cn(
                  "grid grid-cols-1 gap-x-10 mt-10",
                  points.length % 3 === 0 ? "md:grid-cols-3" : "md:grid-cols-2"
                )}
              >
                {points.map((point, i) => (
                  <li key={point.title} className="border-t py-5">
                    <h3 className="font-semibold text-lg flex items-center gap-3">
                      {steps && (
                        <span className="flex-none w-7 h-7 rounded-full bg-primary text-primary-foreground text-sm font-bold flex items-center justify-center">
                          {i + 1}
                        </span>
                      )}
                      {point.href ? (
                        <Link
                          href={point.href}
                          className="inline-flex items-center gap-1 hover:text-primary transition-colors"
                        >
                          {point.title}
                          <ArrowRightIcon className="w-4 h-4" />
                        </Link>
                      ) : steps ? (
                        point.title.replace(stepTitle, "")
                      ) : (
                        point.title
                      )}
                    </h3>
                    <p className="text-muted-foreground mt-2">
                      {point.description}
                    </p>
                  </li>
                ))}
              </List>
            )}
          </section>
        );
      })}

      <HomeFAQSection faqs={content.faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
