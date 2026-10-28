import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import TemplateAppPickerCard from "@/components/app/TemplateAppPickerCard";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { fetchBundle, loginUrlTo } from "@/lib/api/client";
import { useUserQuery } from "@/lib/api/queries";
import { Template, templatePath } from "@/lib/flow/templates";
import { useBundleId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";
import { isExport } from "@/lib/env/client";
import { ogImagePath, templateImageVersion } from "@/lib/og/path";
import { BundleGetResponse } from "@/lib/types/wire.gen";
import { GetStaticPaths, GetStaticProps } from "next";

type Props = {
  bundle: BundleGetResponse | null;
  templateIconSvgs?: TemplateIconSvgs;
};

// On kite.onl the page is rendered on the first visit and again 15 minutes
// later, so link previews show the bundle. The export for self-hosting has no
// server for that, so it has one page for all bundles, which loads them in the
// browser.
export const getStaticPaths: GetStaticPaths | undefined = isExport
  ? undefined
  : () => ({ paths: [], fallback: "blocking" });

export const getStaticProps: GetStaticProps<Props> | undefined = isExport
  ? undefined
  : async ({ params }) => {
      // Bundles that were taken down, or fail to load, are loaded in the
      // browser, where their owner can still see them.
      const bundle = await fetchBundle(String(params?.bundleId)).catch(
        () => null
      );
      // A listed bundle's link leads to its page in the marketplace. It's not
      // permanent, the bundle can leave the marketplace again.
      if (bundle?.listed && !bundle.official) {
        return {
          redirect: { destination: templatePath(bundle.id), permanent: false },
          revalidate: 60,
        };
      }
      return {
        props: {
          bundle,
          templateIconSvgs: templateIconSvgs([bundle?.listing?.icon]),
        },
        revalidate: 60,
      };
    };

// BundlePage is the page of a bundle a user made, which anyone with its link
// can see. Unlisted bundles aren't indexed.
export default function BundlePage({ bundle }: Partial<Props>) {
  const bundleId = useBundleId();
  const template = useTemplate(bundleId, bundle);

  return (
    <HomeLayout
      title={template ? `${template.name} Template` : "Template"}
      description={template?.description}
      image={
        template
          ? ogImagePath("bundles", bundleId, templateImageVersion(template))
          : undefined
      }
      embedButtons={[
        { label: "Use this template", path: `/bundles/${bundleId}#use` },
        { label: "Browse templates", path: "/marketplace" },
      ]}
      noindex
    >
      <div className="container max-w-5xl py-12">
        {template ? (
          <TemplateDetails
            template={template}
            aside={<AddBundleCard template={template} />}
          />
        ) : (
          template === null && (
            <p className="text-muted-foreground text-sm">
              This template doesn&apos;t exist anymore.
            </p>
          )
        )}
      </div>
      <HomeFooter />
    </HomeLayout>
  );
}

// AddBundleCard picks the app to add the bundle to, after logging in.
function AddBundleCard({ template }: { template: Template }) {
  const { data: user } = useUserQuery();

  if (user && !user.success) {
    return (
      <Card id="use" className="p-5 scroll-mt-20">
        <div className="font-semibold mb-1">Add to your app</div>
        <p className="text-sm text-muted-foreground mb-4">
          Log in with Discord to add {template.name} to one of your apps.
        </p>
        <Button className="w-full" asChild>
          <a href={loginUrlTo(`/bundles/${template.id}`)}>Log in</a>
        </Button>
      </Card>
    );
  }

  return (
    <TemplateAppPickerCard
      href={(appId) => ({
        pathname: "/apps/[appId]/bundles/[bundleId]",
        query: { appId, bundleId: template.id },
      })}
    />
  );
}
