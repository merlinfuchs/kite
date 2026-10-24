import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import { TemplateInstalledCard } from "@/components/app/TemplateInstalledCard";
import { useAppId, useBundleId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";

// AppBundlePage shows a bundle the user made like a template, so it can be
// added to the app and updated.
export default function AppBundlePage() {
  const appId = useAppId();
  const template = useTemplate(useBundleId());

  return (
    <AppLayout
      title={template ? `${template.name} Bundle` : "Bundle"}
      breadcrumbs={[
        { label: "Bundles", href: "/apps/[appId]/bundles" },
        {
          label:
            template?.name ?? (template === null ? "Unknown Bundle" : "Bundle"),
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
      ) : (
        template === null && (
          <p className="text-muted-foreground text-sm">
            This bundle doesn&apos;t exist.
          </p>
        )
      )}
    </AppLayout>
  );
}

AppBundlePage.getLayout = getAppShellLayout;
