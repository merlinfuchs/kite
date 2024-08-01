import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";
import AppSettingsAppearance from "@/components/app/AppSettingsAppearance";
import AppSettingsCollaborators from "@/components/app/AppSettingsCollaborators";
import AppSettingsCredentials from "@/components/app/AppSettingsCredentials";
import AppSettingsDelete from "@/components/app/AppSettingsDelete";
import AppSettingsDisable from "@/components/app/AppSettingsDisable";
import AppSettingsInvite from "@/components/app/AppSettingsInvite";

const breadcrumbs = [
  {
    label: "App Settings",
  },
];

export default function AnalyticsPage() {
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
        <AppSettingsInvite />
      </div>
      <Separator className="my-8" />
      <div className="grid gap-6">
        <AppSettingsAppearance />
        <AppSettingsCredentials />
        <AppSettingsCollaborators />

        <div className="flex space-x-3 justify-end">
          <AppSettingsDisable />
          <AppSettingsDelete />
        </div>
      </div>
    </AppLayout>
  );
}
