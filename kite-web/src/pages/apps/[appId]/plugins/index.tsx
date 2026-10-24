import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { PluginList } from "@/components/app/PluginList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";

const breadcrumbs = [
  {
    label: "Plugins",
  },
];

export default function AppPluginsPage() {
  const appId = useAppId();

  return (
    <AppLayout title="App Plugins" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Plugins</h1>
        <p className="text-muted-foreground text-sm">
          Plugins are being replaced by{" "}
          <Link
            href={{ pathname: "/apps/[appId]/templates", query: { appId } }}
            className="text-primary hover:underline"
          >
            templates
          </Link>
          , which add the same features as commands and event listeners you can
          change. Your plugins keep working.{" "}
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
