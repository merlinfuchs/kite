import { BigChartCard } from "@/components/app/BigChartCard";
import { ChartCards } from "@/components/app/ChartCards";
import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";
import { useAppStateGuilds } from "@/lib/hooks/api";
import AppStateGuildList from "@/components/app/AppStateGuildList";

const breadcrumbs = [
  {
    label: "Server Explorer",
  },
];

export default function AppStatePage() {
  const guilds = useAppStateGuilds();

  return (
    <AppLayout title="App Analytics" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Server Explorer
        </h1>
        <p className="text-muted-foreground text-sm">
          Explore the servers your app is in.
        </p>
      </div>
      <Separator className="my-4" />
      <div className="space-y-5">
        <AppStateGuildList />
      </div>
    </AppLayout>
  );
}
