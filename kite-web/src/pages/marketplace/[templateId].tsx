import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { fetchBundle } from "@/lib/api/client";
import { isExport } from "@/lib/env/client";
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

type Props = {
  bundle: BundleGetResponse;
  templateIconSvgs: TemplateIconSvgs;
};

// On kite.onl the page of a template is rendered on its first visit, and is
// checked again every minute. The export for self-hosting has no templates, so
// it has none.
export const getStaticPaths: GetStaticPaths = () => ({
  paths: [],
  fallback: isExport ? false : "blocking",
});

export const getStaticProps: GetStaticProps<Props> = async ({ params }) => {
  const bundle = await fetchBundle(String(params?.templateId));
  // Templates of users are only here while they're listed.
  if (!bundle || !templateFromBundle(bundle)?.listed) {
    return { notFound: true, revalidate: 60 };
  }
  return {
    props: {
      bundle,
      templateIconSvgs: templateIconSvgs([bundle.listing?.icon]),
    },
    revalidate: 60,
  };
};

export default function TemplatePage({ bundle }: Props) {
  const template = templateFromBundle(bundle)!;
  return (
    <HomeLayout
      title={`${template.name} Template for Discord Bots`}
      description={`${template.description} ${
        bundle.official
          ? "A free Kite template."
          : "A free template made with Kite."
      }`}
      // The questions every template users made has aren't structured data,
      // which would repeat them on every page.
      jsonLd={template.faq.length ? [faqJsonLd(template.faq)] : []}
      image={ogImagePath(
        bundle.official ? "templates" : "bundles",
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
