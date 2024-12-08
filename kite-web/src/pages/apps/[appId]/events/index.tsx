import AppLayout from "@/components/app/AppLayout";
import EventListenerList from "@/components/app/EventListerList";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Event Listeners",
  },
];

export default function AppEventsPage() {
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
      <EventListenerList />
    </AppLayout>
  );
}
