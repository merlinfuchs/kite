import AppLayout from "@/components/app/AppLayout";
import AppIntegrationList from "@/components/app/AppIntegrationList";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Integrations",
  },
];

export default function AppIntegrationsPage() {
  return (
    <AppLayout title="Integrations" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Integrations</h1>
        <p className="text-muted-foreground text-sm">
          Enable other services to use their blocks in your flows. Your key
          stays with Kite and only goes to the service it belongs to.
        </p>
      </div>
      <Separator className="my-8" />
      <AppIntegrationList />
    </AppLayout>
  );
}
