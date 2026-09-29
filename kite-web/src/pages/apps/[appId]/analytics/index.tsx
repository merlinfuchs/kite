import AppLayout from "@/components/app/AppLayout";
import UsageAnalytics from "@/components/app/UsageAnalytics";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Analytics",
  },
];

export default function AppAnalyticsPage() {
  return (
    <AppLayout title="Analytics" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Analytics</h1>
        <p className="text-muted-foreground text-sm">
          See how often your commands, events and messages run and how many
          credits they use.
        </p>
      </div>
      <Separator className="my-4" />
      <UsageAnalytics />
    </AppLayout>
  );
}
