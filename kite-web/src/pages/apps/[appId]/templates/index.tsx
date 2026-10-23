import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import { TemplateList } from "@/components/app/TemplateList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { getTemplate } from "@/lib/flow/templates";
import { useRouter } from "next/router";

// A template opens on this page with ?template=<id>, so it doesn't need its
// own route.
export default function AppTemplatesPage() {
  const router = useRouter();

  const templateId = router.query.template as string | undefined;
  const template = templateId ? getTemplate(templateId) : undefined;

  if (template) {
    return (
      <AppLayout
        title={`${template.name} Template`}
        breadcrumbs={[
          { label: "Templates", href: "/apps/[appId]/templates" },
          { label: template.name },
        ]}
      >
        <TemplateDetails template={template} />
      </AppLayout>
    );
  }

  return (
    <AppLayout title="App Templates" breadcrumbs={[{ label: "Templates" }]}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Templates</h1>
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
      <Separator className="my-8" />
      <TemplateList />
    </AppLayout>
  );
}

AppTemplatesPage.getLayout = getAppShellLayout;
