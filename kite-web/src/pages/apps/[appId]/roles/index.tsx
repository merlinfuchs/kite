import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import AppStateGuildExplorer from "@/components/app/AppStateGuildExplorer";
import { Separator } from "@/components/ui/separator";

const breadcrumbs = [
  {
    label: "Role Explorer",
  },
];

export default function AppRolesPage() {
  return (
    <AppLayout title="Role Explorer" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Role Explorer
        </h1>
        <p className="text-muted-foreground text-sm">
          Explore the roles of a server your app is in and what they can do.
        </p>
      </div>
      <Separator className="my-4" />
      <div className="space-y-5">
        <AppStateGuildExplorer view="roles" />
      </div>
    </AppLayout>
  );
}

AppRolesPage.getLayout = getAppShellLayout;
