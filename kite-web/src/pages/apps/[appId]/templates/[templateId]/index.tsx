import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import { TemplateInstalledCard } from "@/components/app/TemplateInstalledCard";
import { useAppId, useTemplateId } from "@/lib/hooks/params";
import { useTemplate } from "@/lib/hooks/templates";

export default function AppTemplatePage() {
  const appId = useAppId();
  const template = useTemplate(useTemplateId());

  return (
    <AppLayout
      title={template ? `${template.name} Template` : "Template"}
      breadcrumbs={[
        { label: "Templates", href: "/apps/[appId]/templates" },
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
          backHref={{ pathname: "/apps/[appId]/templates", query: { appId } }}
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
            This template doesn&apos;t exist.
          </p>
        )
      )}
    </AppLayout>
  );
}

AppTemplatePage.getLayout = getAppShellLayout;
