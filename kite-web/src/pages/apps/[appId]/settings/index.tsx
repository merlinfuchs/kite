import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";
import AppSettingsAppearance from "@/components/app/AppSettingsAppearance";
import AppSettingsCollaborators from "@/components/app/AppSettingsCollaborators";
import AppSettingsCredentials from "@/components/app/AppSettingsCredentials";
import AppSettingsDelete from "@/components/app/AppSettingsDelete";
import AppSettingsPresence from "@/components/app/AppSettingsPresence";
import AppSettingsControls from "@/components/app/AppSettingsControls";

const breadcrumbs = [
  {
    label: "App Settings",
  },
];

export default function AppSettingsPage() {
  return (
    <AppLayout title="App Settings" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row justify-between items-end space-y-5 md:space-y-0">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            App Settings
          </h1>
          <p className="text-muted-foreground text-sm">
            Configure your app settings here. This is where you can manage your
            collaborators and other app settings.
          </p>
        </div>
      </div>
      <Separator className="my-8" />
      <div className="grid gap-6">
        <AppSettingsControls />
        <AppSettingsAppearance />
        <AppSettingsPresence />
        <AppSettingsCredentials />
        <AppSettingsCollaborators />
        <AppSettingsDelete />
      </div>
    </AppLayout>
  );
}
