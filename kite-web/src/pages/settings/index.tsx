import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import KiteSettingsView from "@/components/settings/KiteSettingsView";
import { Separator } from "@/components/ui/separator";
import { useRouter } from "next/router";

const breadcrumbs = [
  {
    label: "Kite Settings",
  },
];

export default function KiteSettingsPage() {
  const router = useRouter();
  const appId = router.query.appId as string | undefined;
  const backHref = appId ? `/apps/${appId}` : "/apps";

  return (
    <AppLayout
      title="Kite Settings"
      breadcrumbs={breadcrumbs}
      backHref={backHref}
    >
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end space-y-5 md:space-y-0">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            Kite Settings
          </h1>
          <p className="text-muted-foreground text-sm">
            Configure your settings for Kite.
          </p>
        </div>
      </div>
      <Separator className="my-5 md:my-8" />
      <KiteSettingsView />
    </AppLayout>
  );
}

KiteSettingsPage.getLayout = getAppShellLayout;
