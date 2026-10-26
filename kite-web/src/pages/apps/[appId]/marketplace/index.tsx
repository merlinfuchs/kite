import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateList } from "@/components/app/TemplateList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { useBundleInstalls } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useTemplates } from "@/lib/hooks/templates";

const breadcrumbs = [
  {
    label: "Marketplace",
  },
];

export default function AppMarketplacePage() {
  const appId = useAppId();
  const installs = useBundleInstalls();
  const templates = useTemplates();

  return (
    <AppLayout title="App Marketplace" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row md:items-start justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            Marketplace
          </h1>
          <p className="text-muted-foreground text-sm">
            Ready-made commands and event listeners you can add to your app and
            change in the editor like anything you built yourself.{" "}
            <a
              href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/templates`}
              target="_blank"
              className="text-primary hover:underline"
            >
              Learn More
            </a>
          </p>
        </div>
      </div>
      <Separator className="my-8" />
      {templates && (
        <TemplateList
          templates={templates}
          href={(t) => ({
            pathname: "/apps/[appId]/marketplace/[templateId]",
            query: { appId, templateId: t.id },
          })}
          added={(t) => !!installs?.some((i) => i?.bundle_id === t.id)}
        />
      )}
    </AppLayout>
  );
}

AppMarketplacePage.getLayout = getAppShellLayout;
