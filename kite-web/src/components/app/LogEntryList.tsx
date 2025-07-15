"use client";

import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import {
  ChevronDown,
  MailPlusIcon,
  RefreshCwIcon,
  SatelliteDishIcon,
  SquareSlash,
  SquareSlashIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
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
import { useLogEntriesQuery } from "@/lib/api/queries";
import { useAppEntities, useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { LogEntry } from "@/lib/types/wire.gen";
import { useMemo, useState } from "react";
import LogLevelBadge from "./LogLevelBadge";

const logLevels = ["debug", "info", "warn", "error"] as const;

export const columns: ColumnDef<{
  level: string;
  message: string;
  source_id: string | null;
  created_at: string;
}>[] = [
  {
    accessorKey: "level",
    header: "Level",

    cell: ({ row }) => {
      const level = row.getValue<string>("level");

      return <LogLevelBadge level={level} />;
    },
  },
  {
    accessorKey: "message",
    header: "Message",
    cell: ({ row }) => <div>{row.getValue("message")}</div>,
  },
  {
    accessorKey: "source_id",
    header: "Source",

    cell: function SourceCell({ row }) {
      const sourceID = row.getValue<string>("source_id");

      const entities = useAppEntities();

      const entity = useMemo(() => {
        if (!sourceID) return null;
        return entities?.find((entity) => entity!.id === sourceID);
      }, [entities, sourceID]);

      if (!entity) return null;

      return (
        <div className="flex items-center gap-1.5 text-foreground/80">
          {entity.type === "command" ? (
            <SquareSlashIcon className="h-4 w-4" />
          ) : entity.type === "event_listener" ? (
            <SatelliteDishIcon className="h-4 w-4" />
          ) : entity.type === "message" ? (
            <MailPlusIcon className="h-4 w-4" />
          ) : (
            <SquareSlash className="h-4 w-4" />
          )}
          <div>{entity?.name ?? sourceID}</div>
        </div>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: () => <div className="text-right">Timestamp</div>,
    cell: ({ row }) => {
      const date = new Date(row.getValue("created_at"));

      let formatted: string;
      if (date.getDay() === new Date().getDay()) {
        formatted = date.toLocaleTimeString("en-US");
      } else {
        formatted = date.toLocaleString("en-US");
      }

      return (
        <div className="text-right font-light text-nowrap">{formatted}</div>
      );
    },
  },
];

export default function LogEntryList({
  commandId,
  eventId,
  messageId,
  disableFilters,
  disablePagination,
}: {
  commandId?: string;
  eventId?: string;
  messageId?: string;
  disableFilters?: boolean;
  disablePagination?: boolean;
}) {
  const query = useLogEntriesQuery(useAppId(), {
    commandId,
    eventId,
    messageId,
  });
  const data = useResponseData(query);

  const [enabledLevels, setEnabledLevels] = useState<string[]>([
    "debug",
    "info",
    "warn",
    "error",
  ]);

  const logEntries = useMemo(() => {
    const entries = (data ?? []) as LogEntry[];

    return entries
      .filter((entry) => enabledLevels.includes(entry!.level))
      .map((entry) => ({
        level: entry.level,
        message: entry.message,
        source_id:
          entry.command_id ?? entry.event_listener_id ?? entry.message_id,
        created_at: entry.created_at,
      }));
  }, [data, enabledLevels]);

  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  const table = useReactTable({
    data: logEntries,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  return (
    <div className="w-full overflow-x-auto lg:overflow-visible">
      {!disableFilters && (
        <div className="flex items-center py-4">
          <Input
            placeholder="Filter logs..."
            value={
              (table.getColumn("message")?.getFilterValue() as string) ?? ""
            }
            onChange={(event) =>
              table.getColumn("message")?.setFilterValue(event.target.value)
            }
            className="max-w-sm mr-2"
          />
          <Button
            variant="outline"
            size="icon"
            className="ml-auto mr-2 flex-none"
            onClick={() => query.refetch()}
          >
            <RefreshCwIcon className="h-4 w-4" />
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline">
                Levels <ChevronDown className="ml-2 h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {logLevels.map((level) => {
                return (
                  <DropdownMenuCheckboxItem
                    key={level}
                    className="capitalize"
                    checked={enabledLevels.includes(level)}
                    onCheckedChange={(value) => {
                      setEnabledLevels((prev) =>
                        value
                          ? [...prev, level]
                          : prev.filter((l) => l !== level)
                      );
                    }}
                  >
                    {level}
                  </DropdownMenuCheckboxItem>
                );
              })}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      )}
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
                  No logs.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      {!disablePagination && (
        <div className="flex items-center justify-end space-x-2 py-4">
          <div className="flex-1 text-sm text-muted-foreground">
            {table.getFilteredRowModel().rows.length} of{" "}
            {table.getRowModel().rows.length} log(s) shown.
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
      )}
    </div>
  );
}
