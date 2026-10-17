import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateList } from "@/components/app/TemplateList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Templates",
  },
];

export default function AppTemplatesPage() {
  return (
    <AppLayout title="App Templates" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Templates</h1>
        <p className="text-muted-foreground text-sm">
          Select any of the templates below to get started. Templates help you
          build your app faster and can contain commands, event listeners,
          message templates, and more.{" "}
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
