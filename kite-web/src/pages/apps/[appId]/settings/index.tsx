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
import { useAppRole } from "@/lib/hooks/api";
import { canManageSettings } from "@/lib/collaborators";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { LockIcon } from "lucide-react";

const breadcrumbs = [
  {
    label: "App settings",
  },
];

export default function AppSettingsPage() {
  const role = useAppRole();
  // Developers can't change the settings, but see who else works on the app
  // and can leave it.
  const isDeveloper = role !== undefined && !canManageSettings(role);

  return (
    <AppLayout title="App settings" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end space-y-5 md:space-y-0">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            App settings
          </h1>
          <p className="text-muted-foreground text-sm">
            Configure your app settings here. This is where you can manage your
            collaborators and other app settings.{" "}
            <a
              href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/managing-your-app#settings`}
              target="_blank"
              className="text-primary hover:underline"
            >
              Learn more
            </a>
          </p>
        </div>
      </div>
      <Separator className="my-8" />
      <div className="grid grid-cols-1 gap-6">
        {isDeveloper ? (
          <Alert>
            <LockIcon className="h-4 w-4" />
            <AlertTitle>You&apos;re a developer of this app</AlertTitle>
            <AlertDescription>
              Developers can edit commands, event listeners, message templates,
              and everything else. Only admins and the owner can change the
              app&apos;s settings.
            </AlertDescription>
          </Alert>
        ) : (
          <>
            <AppSettingsControls />
            <AppSettingsAppearance />
            <AppSettingsPresence />
            <AppSettingsCredentials />
          </>
        )}
        <AppSettingsCollaborators />
        <AppSettingsDelete />
      </div>
    </AppLayout>
  );
}

AppSettingsPage.getLayout = getAppShellLayout;
