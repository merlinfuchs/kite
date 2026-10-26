import { TemplateDetails } from "@/components/app/TemplateDetails";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { fetchBundle, loginUrlTo } from "@/lib/api/client";
import { useUserQuery } from "@/lib/api/queries";
import { Template, templatePath } from "@/lib/flow/templates";
import { useApps } from "@/lib/hooks/api";
import { useBundleId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";
import { isExport } from "@/lib/env/client";
import { ogImagePath } from "@/lib/og/path";
import { BundleGetResponse } from "@/lib/types/wire.gen";
import { GetStaticPaths, GetStaticProps } from "next";
import Link from "next/link";

type Props = { bundle: BundleGetResponse | null };

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
      if (bundle?.listing && !bundle.official) {
        return {
          redirect: { destination: templatePath(bundle.id), permanent: false },
          revalidate: 60,
        };
      }
      return {
        props: { bundle },
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
          ? ogImagePath("bundles", bundleId, template.version)
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

// AddBundleCard picks the app to add the bundle to, after logging in. The
// button of the Discord preview links to it.
function AddBundleCard({ template }: { template: Template }) {
  const { data: user } = useUserQuery();
  const apps = useApps();

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

  const list = apps?.flatMap((a) => (a ? [a] : []));

  return (
    <Card id="use" className="p-5 scroll-mt-20">
      <div className="font-semibold mb-1">Add to your app</div>
      <p className="text-sm text-muted-foreground mb-4">
        {list?.length
          ? "Choose the app. You can fill in its parameters and choose what to include there."
          : "Create an app first, then come back to this page."}
      </p>
      <div className="flex flex-col gap-2">
        {list?.map((app) => (
          <Button key={app.id} variant="outline" className="w-full" asChild>
            <Link
              href={{
                pathname: "/apps/[appId]/bundles/[bundleId]",
                query: { appId: app.id, bundleId: template.id },
              }}
            >
              {app.name}
            </Link>
          </Button>
        ))}
        {list && !list.length && (
          <Button className="w-full" asChild>
            <Link href="/apps">Create an app</Link>
          </Button>
        )}
      </div>
    </Card>
  );
}
