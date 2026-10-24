import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  useReactTable,
} from "@tanstack/react-table";
import {
  CopyIcon,
  HashIcon,
  LogOutIcon,
  MoreHorizontal,
  ShieldIcon,
  UsersIcon,
} from "lucide-react";

import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
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
import { useAppStateGuilds, useResponseData } from "@/lib/hooks/api";
import { useAppStateStatusQuery } from "@/lib/api/queries";
import { Guild } from "@/lib/types/wire.gen";
import { useCallback, useMemo, useState } from "react";
import { useAppStateGuildLeaveMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { useRouter } from "next/router";

export const columns: ColumnDef<Guild>[] = [
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
    header: "Name",
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
  },
  {
    accessorKey: "created_at",
    header: () => <div>Created At</div>,
    cell: ({ row }) => (
      <div>{new Date(row.getValue("created_at")).toLocaleString()}</div>
    ),
  },
  {
    id: "actions",
    enableHiding: false,
    cell: function RowActionCell({ row }) {
      const guild = row.original;
      const router = useRouter();
      const appId = useAppId();

      const [leaveOpen, setLeaveOpen] = useState(false);
      const leaveMutation = useAppStateGuildLeaveMutation(appId);

      const openExplorer = useCallback(
        (view: "channels" | "members" | "roles") => {
          router.push({
            pathname: `/apps/[appId]/${view}`,
            query: { appId, guild: guild.id },
          });
        },
        [router, appId, guild.id]
      );

      const handleLeave = useCallback(() => {
        leaveMutation.mutate(guild.id, {
          onSuccess: (res) => {
            if (res.success) {
              toast.success(`Left ${guild.name}`);
              setLeaveOpen(false);
            } else {
              toast.error(`Failed to leave server: ${res.error.message}`);
            }
          },
        });
      }, [leaveMutation, guild.id, guild.name]);

      return (
        <>
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Actions</DropdownMenuLabel>
              <DropdownMenuItem
                onSelect={() => openExplorer("channels")}
                className="cursor-pointer"
              >
                <HashIcon className="mr-2 h-4 w-4" />
                View channels
              </DropdownMenuItem>
              <DropdownMenuItem
                onSelect={() => openExplorer("members")}
                className="cursor-pointer"
              >
                <UsersIcon className="mr-2 h-4 w-4" />
                View members
              </DropdownMenuItem>
              <DropdownMenuItem
                onSelect={() => openExplorer("roles")}
                className="cursor-pointer"
              >
                <ShieldIcon className="mr-2 h-4 w-4" />
                View roles
              </DropdownMenuItem>
              <DropdownMenuItem
                onSelect={() => {
                  navigator.clipboard.writeText(guild.id);
                  toast.success("Server ID copied to clipboard");
                }}
                className="cursor-pointer"
              >
                <CopyIcon className="mr-2 h-4 w-4" />
                Copy server ID
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onSelect={() => setLeaveOpen(true)}
                className="cursor-pointer text-destructive focus:text-destructive"
              >
                <LogOutIcon className="mr-2 h-4 w-4" />
                Leave server
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <AlertDialog
            open={leaveOpen}
            onOpenChange={(open) =>
              !leaveMutation.isPending && setLeaveOpen(open)
            }
          >
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Leave {guild.name}?</AlertDialogTitle>
                <AlertDialogDescription>
                  Your app will leave this server and stop responding there. To
                  add it back, someone with Manage Server permission has to
                  invite it again.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel disabled={leaveMutation.isPending}>
                  Cancel
                </AlertDialogCancel>
                <Button
                  variant="destructive"
                  onClick={handleLeave}
                  disabled={leaveMutation.isPending}
                >
                  {leaveMutation.isPending ? "Leaving..." : "Leave server"}
                </Button>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </>
      );
    },
  },
];

export default function AppStateGuildList() {
  const guilds = useAppStateGuilds();
  const appStatus = useResponseData(useAppStateStatusQuery(useAppId()));

  const tableData = useMemo(() => (guilds ?? []) as Guild[], [guilds]);

  const table = useReactTable({
    data: tableData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        <Input
          placeholder="Filter servers..."
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn("name")?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
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
                <TableRow key={row.id}>
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
    </div>
  );
}
