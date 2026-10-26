import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { SourceAppCard } from "@/components/app/BundleManageCards";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import { TemplateInstalledCard } from "@/components/app/TemplateInstalledCard";
import { useUserBundles } from "@/lib/hooks/api";
import { useAppId, useTemplateId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";

export default function AppTemplatePage() {
  const appId = useAppId();
  const templateId = useTemplateId();
  const template = useTemplate(templateId);
  // The app a listed bundle of the user is made from can't add it.
  const own = useUserBundles()?.find((b) => b?.id === templateId);
  const fromHere = own?.source_app_id === appId;

  return (
    <AppLayout
      title={template ? `${template.name} Template` : "Template"}
      breadcrumbs={[
        { label: "Marketplace", href: "/apps/[appId]/marketplace" },
        {
          label:
            template?.name ??
            (template === null ? "Unknown Template" : "Template"),
        },
      ]}
    >
      {template ? (
        <TemplateDetails
          template={template}
          backHref={{ pathname: "/apps/[appId]/marketplace", query: { appId } }}
          aside={
            own && fromHere ? (
              <SourceAppCard bundle={own} />
            ) : (
              <div className="space-y-5">
                <TemplateInstalledCard template={template} />
                <TemplateInstallCard template={template} />
              </div>
            )
          }
        />
      ) : (
        template === null && (
          <p className="text-muted-foreground text-sm">
            This template doesn&apos;t exist.
          </p>
        )
      )}
    </AppLayout>
  );
}

AppTemplatePage.getLayout = getAppShellLayout;
