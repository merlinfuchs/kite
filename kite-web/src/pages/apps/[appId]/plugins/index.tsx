import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { PluginList } from "@/components/app/PluginList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Plugins",
  },
];

export default function AppPluginsPage() {
  return (
    <AppLayout title="App Plugins" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Plugins</h1>
        <p className="text-muted-foreground text-sm">
          Plugins are used to extend the functionality of your app without
          having to build things yourself. Plugins work separately from your own
          commands, events, and message templates.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/plugins`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-8" />
      <PluginList />
    </AppLayout>
  );
}

AppPluginsPage.getLayout = getAppShellLayout;
