import AppEmptyPlaceholder from "@/components/app/AppEmptyPlaceholder";
import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Event Listeners",
  },
];

export default function AppCommandsPage() {
  return (
    <AppLayout title="Event Listeners" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Event Listeners
        </h1>
        <p className="text-muted-foreground text-sm">
          Listen for events from your app and take actions based on them.
        </p>
      </div>
      <Separator className="my-8" />
      <AppEmptyPlaceholder
        title="Under construction"
        description="This feature is not yet available. Please check back later."
      />
    </AppLayout>
  );
}
