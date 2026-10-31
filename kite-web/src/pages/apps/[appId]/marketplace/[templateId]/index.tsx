import AppLayout from "@/components/app/AppLayout";
import AppTemplateAside from "@/components/app/AppTemplateAside";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { useAppId, useTemplateId } from "@/lib/hooks/params";
import { useTemplate, useTemplates } from "@/lib/hooks/templates";

export default function AppTemplatePage() {
  const appId = useAppId();
  const templateId = useTemplateId();
  const template = useTemplate(templateId);
  const templates = useTemplates();

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
          related={
            templates && {
              templates,
              href: (t) => ({
                pathname: "/apps/[appId]/marketplace/[templateId]",
                query: { appId, templateId: t.id },
              }),
            }
          }
          aside={<AppTemplateAside template={template} />}
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
