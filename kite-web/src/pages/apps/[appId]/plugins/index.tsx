import AppLayout from "@/components/app/AppLayout";
import { BigChartCard } from "@/components/app/BigChartCard";
import { ChartCards } from "@/components/app/ChartCards";
import { PluginList } from "@/components/app/PluginList";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Modules",
  },
];

export default function AppPluginsPage() {
  return (
    <AppLayout title="Modules" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Modules</h1>
        <p className="text-muted-foreground text-sm">
          Select any of the modules below to get started.
        </p>
      </div>
      <Separator className="my-8" />
      <PluginList />
    </AppLayout>
  );
}
