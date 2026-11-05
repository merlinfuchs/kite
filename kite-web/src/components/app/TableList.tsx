import { Button } from "@/components/ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import TableCreateDialog from "./TableCreateDialog";
import { DatabaseIcon } from "lucide-react";
import { useTablesQuery } from "@/lib/api/queries";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import { useRouter } from "next/router";

export default function TableList() {
  const query = useTablesQuery(useAppId());
  const router = useRouter();
  const tables = query.data?.success ? query.data.data : [];
  const create = (
    <TableCreateDialog>
      <Button>Create table</Button>
    </TableCreateDialog>
  );
  if (query.isPending)
    return (
      <div className="text-sm text-muted-foreground">Loading tables...</div>
    );
  if (!tables.length)
    return (
      <AppEmptyPlaceholder
        icon={DatabaseIcon}
        title="No tables yet"
        description="Create a table with typed columns and edit its rows like a database."
        action={create}
      />
    );
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {tables.map(
        (table) =>
          table && (
            <div className="rounded-lg border p-5 space-y-3" key={table.id}>
              <div className="flex items-center justify-between">
                <h2 className="font-semibold">{table.name}</h2>
                <span className="text-sm text-muted-foreground">
                  {table.columns.length} columns
                </span>
              </div>
              <p className="text-sm text-muted-foreground">
                {table.columns.map((column) => column.name).join(", ")}
              </p>
              <Button asChild size="sm" variant="outline">
                <Link
                  href={{
                    pathname: "/apps/[appId]/tables/[tableId]",
                    query: { appId: router.query.appId, tableId: table.id },
                  }}
                >
                  Open table
                </Link>
              </Button>
            </div>
          )
      )}
    </div>
  );
}
