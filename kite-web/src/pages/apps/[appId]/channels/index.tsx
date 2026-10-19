import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import AppStateGuildExplorer from "@/components/app/AppStateGuildExplorer";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Channel Explorer",
  },
];

export default function AppChannelsPage() {
  return (
    <AppLayout title="Channel Explorer" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Channel Explorer
        </h1>
        <p className="text-muted-foreground text-sm">
          Explore the channels of a server your app is in, and what your app can
          do in them.
        </p>
      </div>
      <Separator className="my-4" />
      <div className="space-y-5">
        <AppStateGuildExplorer view="channels" />
      </div>
    </AppLayout>
  );
}

AppChannelsPage.getLayout = getAppShellLayout;
