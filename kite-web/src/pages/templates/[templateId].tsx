import { TemplateDetails } from "@/components/app/TemplateDetails";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { getTemplate, templatePaths } from "@/lib/flow/templateFiles";
import { Template } from "@/lib/flow/templates";
import { ogImagePath } from "@/lib/og/path";
import { faqJsonLd } from "@/lib/seo";
import { TemplateInstalls, withTemplateInstalls } from "@/lib/templateInstalls";
import { GetStaticPaths, GetStaticProps } from "next";
import Link from "next/link";

type Props = { templateId: string; installs: TemplateInstalls };

// Every template is exported as its own page, so crawlers see its content
// without running JavaScript.
export const getStaticPaths: GetStaticPaths = () => ({
  paths: templatePaths(),
  fallback: false,
});

export const getStaticProps: GetStaticProps<Props> = ({ params }) =>
  withTemplateInstalls({ templateId: String(params?.templateId) });

export default function TemplatePage({ templateId, installs }: Props) {
  const template = getTemplate(templateId)!;

  return (
    <HomeLayout
      title={`${template.name} Template for Discord Bots`}
      description={`${template.description} A free Kite template.`}
      jsonLd={[faqJsonLd(template.faq)]}
      image={ogImagePath("templates", template.id, template.version)}
    >
      <div className="container max-w-5xl py-12">
        <TemplateDetails
          template={template}
          installs={installs?.[template.id]}
          backHref="/templates"
          backOnMobile
          aside={<TemplateGetStartedCard template={template} />}
        />
      </div>

      <HomeFAQSection faqs={template.faq} className="max-w-5xl" />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

function TemplateGetStartedCard({ template }: { template: Template }) {
  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">Use this template</div>
      <p className="text-sm text-muted-foreground mb-4">
        Create your bot in Kite for free, then add {template.name} from
        Templates in your app and change anything you like in the editor.
      </p>
      <Button className="w-full" asChild>
        <Link href="/apps">Get started</Link>
      </Button>
    </Card>
  );
}
