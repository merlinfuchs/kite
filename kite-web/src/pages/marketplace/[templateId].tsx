import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { fetchBundle } from "@/lib/api/client";
import { fetchInstallCount, withApiData } from "@/lib/apiData";
import { isExport } from "@/lib/env/client";
import { getTemplate, templatePaths } from "@/lib/flow/templateFiles";
import {
  Template,
  templateFromBundle,
  templatePath,
} from "@/lib/flow/templates";
import { ogImagePath, templateImageVersion } from "@/lib/og/path";
import { faqJsonLd } from "@/lib/seo";
import { BundleGetResponse } from "@/lib/types/wire.gen";
import { GetStaticPaths, GetStaticProps } from "next";
import Link from "next/link";

// bundle is set for the bundles users listed, null for official templates.
type Props = {
  templateId: string;
  installs: number | null;
  bundle: BundleGetResponse | null;
  templateIconSvgs?: TemplateIconSvgs;
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
  // The content of official templates comes from the files.
  if (getTemplate(templateId)) {
    const installs = await fetchInstallCount(templateId);
    return withApiData({ templateId, installs, bundle: null }, 60);
  }

  // Others are only here while they're listed, and are checked again every
  // minute like the pages of bundles.
  const bundle = await fetchBundle(templateId);
  if (!bundle || bundle.official || !templateFromBundle(bundle)?.listed) {
    return { notFound: true, revalidate: 60 };
  }
  return {
    props: {
      templateId,
      installs: bundle.install_count,
      bundle,
      templateIconSvgs: templateIconSvgs([bundle.listing?.icon]),
    },
    revalidate: 60,
  };
};

export default function TemplatePage({ templateId, installs, bundle }: Props) {
  const template = bundle
    ? templateFromBundle(bundle)!
    : getTemplate(templateId)!;
  return (
    <HomeLayout
      title={`${template.name} Template for Discord Bots`}
      description={`${template.description} ${
        bundle ? "A free template made with Kite." : "A free Kite template."
      }`}
      // The questions every template users made has aren't structured data,
      // which would repeat them on every page.
      jsonLd={template.faq.length ? [faqJsonLd(template.faq)] : []}
      image={ogImagePath(
        bundle ? "bundles" : "templates",
        template.id,
        templateImageVersion(template)
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
          installs={installs ?? undefined}
          backHref="/marketplace"
          backOnMobile
          aside={<TemplateGetStartedCard template={template} />}
        />
      </div>

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
