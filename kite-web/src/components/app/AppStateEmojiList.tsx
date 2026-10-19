import { discordEmojiUrl } from "@/tools/common/utils/discordCdn";
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  useReactTable,
} from "@tanstack/react-table";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useApp, useAppEmojis } from "@/lib/hooks/api";
import { useMemo } from "react";
import Link from "next/link";
import { CopyIcon, ExternalLinkIcon } from "lucide-react";
import { toast } from "sonner";

interface EmojiRow {
  id: string;
  name: string;
  url: string;
  animated: boolean;
}

// The text that shows the emoji in a message, like <:kite:123>.
function emojiMarkdown(emoji: EmojiRow) {
  return `<${emoji.animated ? "a" : ""}:${emoji.name}:${emoji.id}>`;
}

export const columns: ColumnDef<EmojiRow>[] = [
  {
    id: "url",
    cell: ({ row }) => {
      const url = row.original?.url;
      if (!url) {
        return <div className="w-8 h-8 rounded-full bg-muted"></div>;
      }

      return <img src={url} alt="" className="w-8 h-8 rounded-full" />;
    },
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => <div>{row.getValue("name")}</div>,
  },
  {
    accessorKey: "id",
    header: "ID",
    cell: ({ row }) => <div>{row.getValue("id")}</div>,
  },
  {
    accessorKey: "created_at",
    header: "Created At",
    cell: ({ row }) => {
      const id = row.original?.id;

      // Convert discord snowflake to date
      const createdAt = new Date(
        Number(BigInt(id) >> BigInt(22)) + 1420070400000
      );

      return <div>{createdAt.toLocaleString()}</div>;
    },
  },
  {
    id: "actions",
    enableHiding: false,
    cell: ({ row }) => {
      const markdown = emojiMarkdown(row.original);

      return (
        <Button
          variant="ghost"
          size="sm"
          className="flex gap-2 ml-auto"
          title={markdown}
          onClick={() => {
            navigator.clipboard.writeText(markdown);
            toast.success("Emoji markdown copied to clipboard");
          }}
        >
          <CopyIcon className="w-4 h-4" />
          <span>Copy Markdown</span>
        </Button>
      );
    },
  },
];

export default function AppStateEmojiList() {
  const emojis = useAppEmojis();
  const app = useApp();

  const appEmojisUrl = `https://discord.com/developers/applications/${app?.discord_id}/emojis`;

  const tableData = useMemo(
    () =>
      emojis?.map((e) => ({
        id: e!.id,
        name: e!.name,
        url: discordEmojiUrl(e!.id, e!.animated),
        animated: e!.animated,
      })) ?? [],
    [emojis]
  );

  const table = useReactTable({
    data: tableData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  return (
    <div className="w-full">
      <div className="flex items-center py-4 gap-3 justify-between">
        <Input
          placeholder="Filter emojis..."
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn("name")?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
        <Button variant="secondary" className="flex gap-2" asChild>
          <Link href={appEmojisUrl} target="_blank">
            <ExternalLinkIcon className="w-4 h-4" />
            <div>Create Emoji</div>
          </Link>
        </Button>
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
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">
          Showing {table.getRowModel().rows.length} of{" "}
          {table.getFilteredRowModel().rows.length} total emojis.
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
