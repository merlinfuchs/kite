import {
  Column,
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  RowSelectionState,
  SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { ArrowUpDown, MoreHorizontal } from "lucide-react";

import AppStateGuildDetailsDialog from "@/components/app/AppStateGuildDetailsDialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useResponseData } from "@/lib/hooks/api";
import {
  useAppStateGuildsWithCountsQuery,
  useAppStateStatusQuery,
} from "@/lib/api/queries";
import { Guild } from "@/lib/types/wire.gen";
import { useCallback, useMemo, useState } from "react";
import { useAppStateGuildLeaveMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";

function SortableHeader({
  column,
  label,
}: {
  column: Column<Guild>;
  label: string;
}) {
  return (
    <Button
      variant="ghost"
      className="-ml-4 h-8"
      onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
    >
      {label}
      <ArrowUpDown className="ml-2 h-4 w-4" />
    </Button>
  );
}

export const columns: ColumnDef<Guild>[] = [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all servers on this page"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select server"
      />
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    id: "icon_url",
    cell: ({ row }) => {
      const iconUrl = row.original?.icon_url;
      if (!iconUrl) {
        return <div className="w-8 h-8 rounded-full bg-muted"></div>;
      }

      return <img src={iconUrl} alt="" className="w-8 h-8 rounded-full" />;
    },
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: "name",
    header: ({ column }) => <SortableHeader column={column} label="Name" />,
    cell: ({ row }) => <div>{row.getValue("name")}</div>,
    filterFn: (row, _, filterValue) => {
      if (!row.original) return false;

      const inputValue = `${row.original.id} ${row.original.name} ${row.original.description}`;
      return inputValue.toLowerCase().includes(filterValue.toLowerCase());
    },
  },
  {
    accessorKey: "id",
    header: "ID",
    cell: ({ row }) => <div>{row.getValue("id")}</div>,
    enableSorting: false,
  },
  {
    id: "member_count",
    header: ({ column }) => <SortableHeader column={column} label="Members" />,
    cell: ({ row }) => {
      const count = row.original.member_count;
      return <div>{count === null ? "-" : count.toLocaleString()}</div>;
    },
    sortUndefined: "last",
    accessorFn: (guild) => guild.member_count ?? undefined,
  },
  {
    id: "joined_at",
    header: ({ column }) => <SortableHeader column={column} label="Joined" />,
    cell: ({ row }) => {
      const joinedAt = row.original.joined_at;
      return (
        <div>{joinedAt ? new Date(joinedAt).toLocaleDateString() : "-"}</div>
      );
    },
    sortUndefined: "last",
    accessorFn: (guild) => guild.joined_at ?? undefined,
  },
  {
    accessorKey: "created_at",
    header: ({ column }) => <SortableHeader column={column} label="Created" />,
    cell: ({ row }) => (
      <div>{new Date(row.original.created_at).toLocaleDateString()}</div>
    ),
  },
  {
    id: "actions",
    enableHiding: false,
    cell: function RowActionCell({ row }) {
      const guild = row.original;

      const appId = useAppId();
      const leaveMutation = useAppStateGuildLeaveMutation(appId);

      const [detailsOpen, setDetailsOpen] = useState(false);

      const handleCopyId = useCallback(() => {
        navigator.clipboard.writeText(guild.id);
        toast.success("Server ID copied");
      }, [guild.id]);

      const handleLeave = useCallback(() => {
        if (confirm("Are you sure you want the app to leave this server?")) {
          leaveMutation.mutate(guild.id, {
            onSuccess: (res) => {
              if (res.success) {
                toast.success("Server left successfully");
              } else {
                toast.error(`Failed to leave server: ${res.error.message}`);
              }
            },
          });
        }
      }, [leaveMutation, guild.id]);

      return (
        <>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Actions</DropdownMenuLabel>
              <DropdownMenuItem
                onClick={() => setDetailsOpen(true)}
                role="button"
                className="cursor-pointer"
              >
                Owner and permissions
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={handleCopyId}
                role="button"
                className="cursor-pointer"
              >
                Copy server ID
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={handleLeave}
                role="button"
                className="cursor-pointer"
              >
                Leave server
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <AppStateGuildDetailsDialog
            guild={guild}
            open={detailsOpen}
            onOpenChange={setDetailsOpen}
          />
        </>
      );
    },
  },
];

export default function AppStateGuildList() {
  const appId = useAppId();
  const guilds = useResponseData(useAppStateGuildsWithCountsQuery(appId));
  const appStatus = useResponseData(useAppStateStatusQuery(appId));

  const leaveMutation = useAppStateGuildLeaveMutation(appId);

  const [sorting, setSorting] = useState<SortingState>([]);
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const [bulkLeaveOpen, setBulkLeaveOpen] = useState(false);
  const [bulkLeaving, setBulkLeaving] = useState(false);

  const tableData = useMemo(() => (guilds ?? []) as Guild[], [guilds]);

  const table = useReactTable({
    data: tableData,
    columns,
    // Selection is kept by server ID, so it survives sorting and refetches.
    getRowId: (guild) => guild.id,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      rowSelection,
    },
  });

  // Servers hidden by the filter are never left, even if they were selected
  // before it was typed.
  const selectedGuilds = table
    .getFilteredSelectedRowModel()
    .rows.map((row) => row.original);

  const handleBulkLeave = useCallback(async () => {
    setBulkLeaving(true);

    let left = 0;
    let failed = 0;

    // One at a time, to stay clear of Discord's rate limits.
    for (const guild of selectedGuilds) {
      try {
        const res = await leaveMutation.mutateAsync(guild.id);
        if (res.success) {
          left++;
        } else {
          failed++;
        }
      } catch (e) {
        failed++;
      }
    }

    setBulkLeaving(false);
    setBulkLeaveOpen(false);
    setRowSelection({});

    if (failed === 0) {
      toast.success(`Left ${left} server${left === 1 ? "" : "s"}`);
    } else {
      toast.error(
        `Left ${left} server${left === 1 ? "" : "s"}, failed to leave ${failed}`
      );
    }
  }, [selectedGuilds, leaveMutation]);

  return (
    <div className="w-full">
      <div className="flex items-center justify-between gap-3 py-4">
        <Input
          placeholder="Filter servers..."
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn("name")?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
        {selectedGuilds.length > 0 && (
          <Button
            variant="destructive"
            size="sm"
            onClick={() => setBulkLeaveOpen(true)}
          >
            Leave {selectedGuilds.length} selected
          </Button>
        )}
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  {appStatus && !appStatus.online
                    ? "Your app is offline. Start it to see its servers."
                    : "No results."}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">
          Showing {table.getRowModel().rows.length} of{" "}
          {table.getFilteredRowModel().rows.length} total servers.
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>

      <AlertDialog
        open={bulkLeaveOpen}
        onOpenChange={(open) => !bulkLeaving && setBulkLeaveOpen(open)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Leave {selectedGuilds.length} server
              {selectedGuilds.length === 1 ? "" : "s"}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              Your app will leave these servers. To get it back, someone has to
              invite it again.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <ul className="max-h-60 overflow-y-auto rounded-md border px-3 py-2 text-sm space-y-1">
            {selectedGuilds.map((guild) => (
              <li key={guild.id} className="flex justify-between gap-3">
                <span className="truncate">{guild.name}</span>
                <span className="text-muted-foreground font-mono text-xs">
                  {guild.id}
                </span>
              </li>
            ))}
          </ul>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={bulkLeaving}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              disabled={bulkLeaving}
              onClick={(event) => {
                // Keeps the dialog open until every server has been left.
                event.preventDefault();
                handleBulkLeave();
              }}
            >
              {bulkLeaving ? "Leaving..." : "Leave servers"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
