import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { TemplateInstallCard } from "@/components/app/TemplateInstallCard";
import { getTemplate } from "@/lib/flow/templates";
import { useAppId, useTemplateId } from "@/lib/hooks/params";
import { useRouter } from "next/router";

export default function AppTemplatePage() {
  const router = useRouter();
  const appId = useAppId();
  const template = getTemplate(useTemplateId());

  return (
    <AppLayout
      title={template ? `${template.name} Template` : "Template"}
      breadcrumbs={[
        { label: "Templates", href: "/apps/[appId]/templates" },
        { label: template?.name ?? "Unknown Template" },
      ]}
    >
      {template ? (
        <TemplateDetails
          template={template}
          backHref={{ pathname: "/apps/[appId]/templates", query: { appId } }}
          aside={<TemplateInstallCard template={template} />}
        />
      ) : (
        router.isReady && (
          <p className="text-muted-foreground text-sm">
            This template doesn&apos;t exist.
          </p>
        )
      )}
    </AppLayout>
  );
}

AppTemplatePage.getLayout = getAppShellLayout;
