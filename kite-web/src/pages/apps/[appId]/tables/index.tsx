import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import TableList from "@/components/app/TableList";
import TableCreateDialog from "@/components/app/TableCreateDialog";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

export default function TablesPage() {
  return (
    <AppLayout title="Tables" breadcrumbs={[{ label: "Tables" }]}>
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl">Tables</h1>
          <p className="text-sm text-muted-foreground">
            Store structured content in editable tables with typed columns.
          </p>
        </div>
        <TableCreateDialog>
          <Button>Create table</Button>
        </TableCreateDialog>
      </div>
      <Separator className="my-8" />
      <TableList />
    </AppLayout>
  );
}
TablesPage.getLayout = getAppShellLayout;
