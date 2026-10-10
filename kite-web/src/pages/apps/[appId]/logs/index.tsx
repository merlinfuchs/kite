import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import LogEntryList from "@/components/app/LogEntryList";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Logs",
  },
];

export default function AppLogsPage() {
  return (
    <AppLayout title="Logs" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Logs</h1>
        <p className="text-muted-foreground text-sm">
          View the logs for your app. Some logs are produced by Kite itself, but
          you can also add your own logs to your flows.{" "}
          <a
            href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/troubleshooting#logs`}
            target="_blank"
            className="text-primary hover:underline"
          >
            Learn More
          </a>
        </p>
      </div>
      <Separator className="my-4" />
      <div className="space-y-5">
        <LogEntryList />
      </div>
    </AppLayout>
  );
}

AppLogsPage.getLayout = getAppShellLayout;
