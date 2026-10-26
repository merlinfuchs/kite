import { TemplateDetails } from "@/components/app/TemplateDetails";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { fetchBundle } from "@/lib/api/client";
import { fetchStats, Stats, withApiData } from "@/lib/apiData";
import { isExport } from "@/lib/env/client";
import { getTemplate, templatePaths } from "@/lib/flow/templateFiles";
import {
  Template,
  templateFromBundle,
  templatePath,
} from "@/lib/flow/templates";
import { communityTemplateFaq } from "@/lib/flow/templateFaq";
import { ogImagePath } from "@/lib/og/path";
import { faqJsonLd } from "@/lib/seo";
import { BundleGetResponse } from "@/lib/types/wire.gen";
import { GetStaticPaths, GetStaticProps } from "next";
import Link from "next/link";

// bundle is set for the bundles users listed, null for official templates.
type Props = {
  templateId: string;
  stats: Stats;
  bundle: BundleGetResponse | null;
};

// Every official template is exported as its own page, so crawlers see its
// content without running JavaScript. On kite.onl the bundles users listed
// are rendered on their first visit. The export for self-hosting only has the
// official ones.
export const getStaticPaths: GetStaticPaths = () => ({
  paths: templatePaths(),
  fallback: isExport ? false : "blocking",
});

export const getStaticProps: GetStaticProps<Props> = async ({ params }) => {
  const templateId = String(params?.templateId);
  if (getTemplate(templateId)) {
    return withApiData({ templateId, stats: await fetchStats(), bundle: null });
  }

  // Others are only here while they're listed, and are checked again every
  // minute like the pages of bundles.
  const [bundle, stats] = await Promise.all([
    fetchBundle(templateId),
    fetchStats(),
  ]);
  if (!bundle || bundle.official || !templateFromBundle(bundle)?.listed) {
    return { notFound: true, revalidate: 60 };
  }
  return { props: { templateId, stats, bundle }, revalidate: 60 };
};

export default function TemplatePage({ templateId, stats, bundle }: Props) {
  const template = bundle
    ? templateFromBundle(bundle)!
    : getTemplate(templateId)!;
  // The questions every template users made has go after its own, and aren't
  // structured data, which would repeat them on every page.
  const faq = bundle
    ? [...template.faq, ...communityTemplateFaq(template)]
    : template.faq;

  return (
    <HomeLayout
      title={`${template.name} Template for Discord Bots`}
      description={`${template.description} ${
        bundle ? "A free template made with Kite." : "A free Kite template."
      }`}
      jsonLd={template.faq.length ? [faqJsonLd(template.faq)] : []}
      image={ogImagePath(
        bundle ? "bundles" : "templates",
        template.id,
        template.version
      )}
      embedButtons={[
        {
          label: "Use this template",
          path: `${templatePath(template.id)}#use`,
        },
        { label: "All templates", path: "/marketplace" },
      ]}
    >
      <div className="container max-w-5xl py-12">
        <TemplateDetails
          template={template}
          installs={stats?.template_installs?.[template.id]}
          backHref="/marketplace"
          backOnMobile
          aside={<TemplateGetStartedCard template={template} />}
        />
      </div>

      {faq.length > 0 && <HomeFAQSection faqs={faq} className="max-w-5xl" />}
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

function TemplateGetStartedCard({ template }: { template: Template }) {
  return (
    // The button of the Discord preview links here.
    <Card id="use" className="p-5 scroll-mt-20">
      <div className="font-semibold mb-1">Use this template</div>
      <p className="text-sm text-muted-foreground mb-4">
        Create your bot in Kite for free, then add {template.name} from the
        Marketplace in your app and change anything you like in the editor.
      </p>
      <Button className="w-full" asChild>
        <Link href="/apps">Get started</Link>
      </Button>
    </Card>
  );
}
