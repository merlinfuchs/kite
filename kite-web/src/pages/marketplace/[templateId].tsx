import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import TemplateAppPickerCard from "@/components/app/TemplateAppPickerCard";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import MarketplaceBar from "@/components/home/MarketplaceBar";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { fetchBundle, loginUrlTo } from "@/lib/api/client";
import { useUserQuery } from "@/lib/api/queries";
import { isExport } from "@/lib/env/client";
import {
  categoryPath,
  relatedTemplates,
  Template,
  templateFromBundle,
  templatePath,
} from "@/lib/flow/templates";
import { ogImagePath, templateImageVersion } from "@/lib/og/path";
import { faqJsonLd } from "@/lib/seo";
import { fetchBundles } from "@/lib/apiData";
import { Bundle, BundleGetResponse } from "@/lib/types/wire.gen";
import { GetStaticPaths, GetStaticProps } from "next";

type Props = {
  bundle: BundleGetResponse;
  // The ones the page suggests, see relatedTemplates.
  related: Bundle[];
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
  const template = bundle ? templateFromBundle(bundle) : undefined;
  if (!bundle || !template?.listed) {
    return { notFound: true, revalidate: 60 };
  }

  // Only the bundles it suggests, not all of the marketplace.
  const bundles = (await fetchBundles()) ?? [];
  const { byCreator, similar } = relatedTemplates(
    template,
    bundles.flatMap((b) => templateFromBundle(b) ?? [])
  );
  const relatedIds = [...byCreator, ...similar].map((t) => t.id);
  const related = bundles.filter((b) => relatedIds.includes(b.id));

  return {
    props: {
      bundle,
      related,
      templateIconSvgs: templateIconSvgs(
        [bundle, ...related].map((b) => b.listing?.icon)
      ),
    },
    revalidate: 60,
  };
};

export default function TemplatePage({ bundle, related }: Props) {
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
      <MarketplaceBar category={template.category ?? null} />
      <div className="container max-w-5xl py-12">
        <TemplateDetails
          template={template}
          aside={<TemplateGetStartedCard template={template} />}
          related={{
            templates: related.flatMap((b) => templateFromBundle(b) ?? []),
            categoryHref: categoryPath,
            href: (t) => templatePath(t.id),
          }}
        />
      </div>

      <HomeCTASection light />
      <HomeFooter />
    </HomeLayout>
  );
}

// TemplateGetStartedCard lets logged in users add the template to one of their
// apps. The page is rendered on the server for everyone, so others, and
// everyone until the user loads, see how to get started, which logs them in and
// brings them back here.
function TemplateGetStartedCard({ template }: { template: Template }) {
  const { data: user } = useUserQuery();

  if (user?.success) {
    return (
      <TemplateAppPickerCard
        href={(appId) => ({
          pathname: "/apps/[appId]/marketplace/[templateId]",
          query: { appId, templateId: template.id },
        })}
      />
    );
  }

  return (
    // The button of the Discord preview links here.
    <Card id="use" className="p-5 scroll-mt-20">
      <div className="font-semibold mb-1">Use this template</div>
      <p className="text-sm text-muted-foreground mb-4">
        Create your bot in Kite for free, then add {template.name} from the
        Marketplace in your app and change anything you like in the editor.
      </p>
      <Button className="w-full" asChild>
        <a href={loginUrlTo(templatePath(template.id))}>Get started</a>
      </Button>
    </Card>
  );
}
