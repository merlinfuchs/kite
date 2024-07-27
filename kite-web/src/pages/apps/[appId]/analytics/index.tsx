import { BigChartCard } from "@/components/app/BigChartCard";
import { ChartCards } from "@/components/app/ChartCards";
import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "App Analytics",
  },
];

export default function AppAnalyticsPage() {
  return (
    <AppLayout title="App Analytics" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Analytics</h1>
        <p className="text-muted-foreground text-sm">
          Gain insights into your app with analytics. Track your plugin&apos;s
          performance and make data-driven decisions.
        </p>
      </div>
      <Separator className="my-4" />
      <div className="space-y-5">
        <ChartCards />
        <BigChartCard />
      </div>
    </AppLayout>
  );
}
