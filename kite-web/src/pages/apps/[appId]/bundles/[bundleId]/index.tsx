import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import BundleManageCards, {
  SourceAppCard,
} from "@/components/app/BundleManageCards";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import {
  RemoveInstallButton,
  TemplateInstalledCard,
} from "@/components/app/TemplateInstalledCard";
import { Card } from "@/components/ui/card";
import { useBundleInstalls, useUserBundles } from "@/lib/hooks/api";
import { useAppId, useBundleId } from "@/lib/hooks/params";
import { useBundleQuery } from "@/lib/api/queries";
import { useTemplate } from "@/lib/hooks/templates";
import { BundleInstall } from "@/lib/types/wire.gen";

// AppBundlePage shows a bundle like a template, so it can be added to the app
// and updated. The app it's made from has its items already.
export default function AppBundlePage() {
  const appId = useAppId();
  const bundleId = useBundleId();
  const template = useTemplate(bundleId);
  const { data } = useBundleQuery(bundleId);
  const removed =
    data?.success === false && data.error.code === "bundle_removed";
  const install = useBundleInstalls()?.find((i) => i?.bundle_id === bundleId);
  // The user's own bundle is managed here too.
  const own = useUserBundles()?.find((b) => b?.id === bundleId);
  const fromHere = own?.source_app_id === appId;

  return (
    <AppLayout
      title={template ? `${template.name} Template` : "Template"}
      breadcrumbs={[
        { label: "Your Templates", href: "/apps/[appId]/bundles" },
        {
          label:
            template?.name ??
            install?.name ??
            (template === null ? "Unknown Template" : "Template"),
        },
      ]}
    >
      {template ? (
        <TemplateDetails
          template={template}
          backHref={{ pathname: "/apps/[appId]/bundles", query: { appId } }}
          aside={
            <div className="space-y-5">
              {own && fromHere ? (
                <SourceAppCard bundle={own} />
              ) : (
                <>
                  <TemplateInstalledCard template={template} />
                  <TemplateInstallCard template={template} />
                </>
              )}
              {own && <BundleManageCards bundle={own} />}
            </div>
          }
        />
      ) : template === null && install ? (
        <GoneBundleCard install={install} removed={removed} />
      ) : (
        template === null && (
          <p className="text-muted-foreground text-sm">
            This template doesn&apos;t exist anymore.
          </p>
        )
      )}
    </AppLayout>
  );
}

// GoneBundleCard is for a bundle the app added that the user can't see, so
// what it added can still be removed.
function GoneBundleCard({
  install,
  removed,
}: {
  install: BundleInstall;
  removed: boolean;
}) {
  return (
    <Card className="p-5 max-w-md">
      <div className="font-semibold mb-1">{install.name}</div>
      {removed ? (
        <p className="text-sm text-muted-foreground mb-4">
          Kite took this template down for breaking the rules. What it added is
          still in your app and keeps running, so check what it does, and remove
          it if you&apos;re not sure.
        </p>
      ) : (
        <p className="text-sm text-muted-foreground mb-4">
          This template doesn&apos;t exist anymore, whoever made it deleted it.
          What it added is still in your app and works as before, but
          doesn&apos;t get updates.
        </p>
      )}
      <RemoveInstallButton
        install={install}
        name={install.name}
        what="everything it added that your app still has"
      />
    </Card>
  );
}

AppBundlePage.getLayout = getAppShellLayout;
