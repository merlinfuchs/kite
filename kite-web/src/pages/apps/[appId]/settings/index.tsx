import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";
import AppSettingsAppearance from "@/components/app/AppSettingsAppearance";
import AppSettingsCollaborators from "@/components/app/AppSettingsCollaborators";
import AppSettingsCredentials from "@/components/app/AppSettingsCredentials";
import AppSettingsDelete from "@/components/app/AppSettingsDelete";

const breadcrumbs = [
  {
    label: "App Settings",
  },
];

export default function AnalyticsPage() {
  return (
    <AppLayout title="App Settings" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">App Settings</h1>
        <p className="text-muted-foreground text-sm">
          Configure your app settings here. This is where you can manage your
          collaborators and other app settings.
        </p>
      </div>
      <Separator className="my-8" />
      <div className="grid gap-6">
        <AppSettingsAppearance />
        <AppSettingsCredentials />
        <AppSettingsCollaborators />

        <div className="flex justify-end">
          <AppSettingsDelete />
        </div>
      </div>
    </AppLayout>
  );
}
