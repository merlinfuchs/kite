import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import AppStateGuildExplorer from "@/components/app/AppStateGuildExplorer";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Member Explorer",
  },
];

export default function AppMembersPage() {
  return (
    <AppLayout title="Member Explorer" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Member Explorer
        </h1>
        <p className="text-muted-foreground text-sm">
          Explore the members of a server your app is in.
        </p>
      </div>
      <Separator className="my-4" />
      <div className="space-y-5">
        <AppStateGuildExplorer view="members" />
      </div>
    </AppLayout>
  );
}

AppMembersPage.getLayout = getAppShellLayout;
