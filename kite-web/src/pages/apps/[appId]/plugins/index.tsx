import AppLayout from "@/components/app/AppLayout";
import { PluginList } from "@/components/app/PluginList";
import { TemplateList } from "@/components/app/TemplateList";
import { Separator } from "@/components/ui/separator";

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
          Plugins are used to extend the functionality of your app. They can
          contain commands, event listeners, message templates, and more.
        </p>
      </div>
      <Separator className="my-8" />
      <PluginList />
    </AppLayout>
  );
}
