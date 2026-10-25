import { TemplateDetails } from "@/components/app/TemplateDetails";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { loginUrlTo } from "@/lib/api/client";
import { useUserQuery } from "@/lib/api/queries";
import { Template } from "@/lib/flow/templates";
import { useApps } from "@/lib/hooks/api";
import { useBundleId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";
import Link from "next/link";

// BundlePage is the page of a bundle a user published, which anyone with its
// link can see. It's loaded in the browser, unlisted bundles aren't indexed.
export default function BundlePage() {
  const template = useTemplate(useBundleId());

  return (
    <HomeLayout
      title={template ? `${template.name} Bundle` : "Bundle"}
      description={template?.description}
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
              This bundle doesn&apos;t exist or isn&apos;t published anymore.
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
  const apps = useApps();

  if (user && !user.success) {
    return (
      <Card className="p-5">
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
    <Card className="p-5">
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
