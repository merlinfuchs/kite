import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { faqJsonLd } from "@/lib/seo";
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
    points?: { title: string; description: string }[];
  }[];
  faqs: FAQProps[];
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

      {content.sections.map((section) => (
        <section key={section.heading} className="container py-12 md:py-16">
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
      ))}

      <HomeFAQSection faqs={content.faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
