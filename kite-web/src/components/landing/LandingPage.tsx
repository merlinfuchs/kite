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

export interface LandingPageContent {
  title: string;
  description: string;
  heading: string;
  intro: string;
  secondaryAction?: { label: string; href: string };
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

      {content.comparison && (
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
      )}

      <HomeFAQSection faqs={content.faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
