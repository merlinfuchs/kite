import AppEmptyPlaceholder from "@/components/app/AppEmptyPlaceholder";
import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Message Templates",
  },
];

export default function AppMessagesPage() {
  return (
    <AppLayout title="Message Templates" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Message Templates
        </h1>
        <p className="text-muted-foreground text-sm">
          Create message templates that can be used as responses to commands and
          events in your app.
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
