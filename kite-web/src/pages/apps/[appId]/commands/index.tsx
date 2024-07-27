import CommandList from "@/components/app/CommandList";
import AppLayout from "@/components/app/AppLayout";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Commands",
  },
];

export default function AppCommandsPage() {
  return (
    <AppLayout title="Commands" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Commands</h1>
        <p className="text-muted-foreground text-sm">
          Create custom commands for your app to let users interact with it.
        </p>
      </div>
      <Separator className="my-4" />
      <CommandList />
    </AppLayout>
  );
}
