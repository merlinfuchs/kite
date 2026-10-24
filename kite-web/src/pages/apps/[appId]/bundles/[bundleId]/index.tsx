import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import {
  RemoveInstallButton,
  TemplateInstalledCard,
} from "@/components/app/TemplateInstalledCard";
import { Card } from "@/components/ui/card";
import { useBundleInstalls } from "@/lib/hooks/api";
import { useAppId, useBundleId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";
import { BundleInstall } from "@/lib/types/wire.gen";

// AppBundlePage shows a bundle like a template, so it can be added to the app
// and updated.
export default function AppBundlePage() {
  const appId = useAppId();
  const bundleId = useBundleId();
  const template = useTemplate(bundleId);
  const install = useBundleInstalls()?.find((i) => i?.bundle_id === bundleId);

  return (
    <AppLayout
      title={template ? `${template.name} Bundle` : "Bundle"}
      breadcrumbs={[
        { label: "Bundles", href: "/apps/[appId]/bundles" },
        {
          label:
            template?.name ??
            install?.name ??
            (template === null ? "Unknown Bundle" : "Bundle"),
        },
      ]}
    >
      {template ? (
        <TemplateDetails
          template={template}
          backHref={{ pathname: "/apps/[appId]/bundles", query: { appId } }}
          aside={
            <div className="space-y-5">
              <TemplateInstalledCard template={template} />
              <TemplateInstallCard template={template} />
            </div>
          }
        />
      ) : template === null && install ? (
        <GoneBundleCard install={install} />
      ) : (
        template === null && (
          <p className="text-muted-foreground text-sm">
            This bundle doesn&apos;t exist or isn&apos;t published.
          </p>
        )
      )}
    </AppLayout>
  );
}

// GoneBundleCard is for a bundle the app added that the user can't see, so
// what it added can still be removed.
function GoneBundleCard({ install }: { install: BundleInstall }) {
  return (
    <Card className="p-5 max-w-md">
      <div className="font-semibold mb-1">{install.name}</div>
      <p className="text-sm text-muted-foreground mb-4">
        You can&apos;t see this bundle anymore. It was unpublished or deleted,
        or whoever made it hasn&apos;t published it, so only they can update it.
        What it added is still in your app and works as before.
      </p>
      <RemoveInstallButton
        install={install}
        name={install.name}
        what="everything it added that your app still has"
      />
    </Card>
  );
}

AppBundlePage.getLayout = getAppShellLayout;
