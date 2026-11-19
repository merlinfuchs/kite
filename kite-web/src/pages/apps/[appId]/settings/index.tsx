import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import AppSettingsAppearance from "@/components/app/AppSettingsAppearance";
import AppSettingsCollaborators from "@/components/app/AppSettingsCollaborators";
import AppSettingsControls from "@/components/app/AppSettingsControls";
import AppSettingsCredentials from "@/components/app/AppSettingsCredentials";
import AppSettingsDelete from "@/components/app/AppSettingsDelete";
import AppSettingsPresence from "@/components/app/AppSettingsPresence";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { useApp, useSharedBot } from "@/lib/hooks/api";

const breadcrumbs = [
  {
    label: "App settings",
  },
];

export default function AppSettingsPage() {
  // The Kite bot's name, avatar, status, and token aren't the app's, and who
  // can change the app follows Manage Server in the server.
  const sharedBot = useSharedBot();
  // Which settings there are is only known once the app is loaded.
  const loaded = !!useApp();

  return (
    <AppLayout title="App settings" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end space-y-5 md:space-y-0">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            App settings
          </h1>
          {sharedBot ? (
            <p className="text-muted-foreground text-sm">
              Everyone with Manage Server in your server can change the
              templates and these settings.
            </p>
          ) : (
            <p className="text-muted-foreground text-sm">
              Configure your app settings here. This is where you can manage
              your collaborators and other app settings.{" "}
              <a
                href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/managing-your-app#settings`}
                target="_blank"
                className="text-primary hover:underline"
              >
                Learn more
              </a>
            </p>
          )}
        </div>
      </div>
      <Separator className="my-8" />
      <div className="grid grid-cols-1 gap-6">
        <AppSettingsControls />
        {loaded && !sharedBot && (
          <>
            <AppSettingsAppearance />
            <AppSettingsPresence />
            <AppSettingsCredentials />
            <AppSettingsCollaborators />
          </>
        )}
        <AppSettingsDelete />
      </div>
    </AppLayout>
  );
}

AppSettingsPage.getLayout = getAppShellLayout;
