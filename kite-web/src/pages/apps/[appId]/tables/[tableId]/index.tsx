import AppLayout from "@/components/app/AppLayout";
import CustomSelect from "@/components/common/CustomSelect";
import ConfirmDialog from "@/components/common/ConfirmDialog";
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
import { getAppShellLayout } from "@/components/app/AppShell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import {
  useTableRowCreateMutation,
  useTableRowDeleteMutation,
  useTableRowUpdateMutation,
  useTableDeleteMutation,
  useTableUpdateMutation,
  useTableAuditUndoMutation,
} from "@/lib/api/mutations";
import {
  useTableQuery,
  useTableRowsQuery,
  useTableAuditQuery,
  useTablesQuery,
} from "@/lib/api/queries";
import { useAppId, useTableId } from "@/lib/hooks/params";
import { Table, TableColumn, TableRow } from "@/lib/types/wire.gen";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ChevronLeftIcon,
  DatabaseIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react";
import { useMemo, useState } from "react";
import { useRouter } from "next/router";
import { toast } from "sonner";

type Sort = { column: string; direction: "asc" | "desc" };

export default function TablePage() {
  const appId = useAppId();
  const tableId = useTableId();
  const tableQuery = useTableQuery(appId, tableId);
  const rowsQuery = useTableRowsQuery(appId, tableId);
  const tablesQuery = useTablesQuery(appId);
  const router = useRouter();
  const table = tableQuery.data?.success ? tableQuery.data.data : undefined;
  const rows = useMemo(
    () =>
      rowsQuery.data?.success
        ? rowsQuery.data.data.filter((row): row is TableRow => !!row)
        : [],
    [rowsQuery.data]
  );
  const create = useTableRowCreateMutation(appId, tableId);
  const [search, setSearch] = useState("");
  const [sort, setSort] = useState<Sort | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [view, setView] = useState<"data" | "schema" | "history">("data");
  const historyQuery = useTableAuditQuery(appId, tableId);
  const undo = useTableAuditUndoMutation(appId, tableId);

  const visibleRows = useMemo(() => {
    const query = search.trim().toLowerCase();
    const filtered = query
      ? rows.filter(
          (row) =>
            row.id.toLowerCase().includes(query) ||
            Object.values(row.data).some((value) =>
              String(value ?? "")
                .toLowerCase()
                .includes(query)
            )
        )
      : rows;
    if (!sort) return filtered;
    return [...filtered].sort((a, b) => {
      const left = String(a.data[sort.column] ?? "");
      const right = String(b.data[sort.column] ?? "");
      return (
        left.localeCompare(right, undefined, { numeric: true }) *
        (sort.direction === "asc" ? 1 : -1)
      );
    });
  }, [rows, search, sort]);

  if (!table) {
    return (
      <AppLayout
        title="Table"
        breadcrumbs={[{ label: "Tables", href: `/apps/${appId}/tables` }]}
      >
        <div className="text-sm text-muted-foreground">Loading table...</div>
      </AppLayout>
    );
  }

  const allVisibleSelected =
    visibleRows.length > 0 && visibleRows.every((row) => selected.has(row.id));
  const addRow = () => {
    create.mutate(
      {
        data: Object.fromEntries(
          table.columns.map((column) => [column.id, ""])
        ),
      },
      {
        onSuccess: (res) => {
          if (!res.success)
            toast.error(`Failed to add row: ${res.error.message}`);
        },
      }
    );
  };
  const toggleSort = (column: string) => {
    setSort((current) =>
      current?.column === column
        ? current.direction === "asc"
          ? { column, direction: "desc" }
          : null
        : { column, direction: "asc" }
    );
  };
  const toggleAll = (checked: boolean | "indeterminate") => {
    if (checked === true)
      setSelected(new Set(visibleRows.map((row) => row.id)));
    else setSelected(new Set());
  };

  return (
    <AppLayout
      title={table.name}
      breadcrumbs={[
        { label: "Tables", href: `/apps/${appId}/tables` },
        { label: table.name },
      ]}
    >
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-start gap-3">
          <div className="rounded-lg border bg-muted/40 p-2.5">
            <DatabaseIcon className="h-5 w-5 text-muted-foreground" />
          </div>
          <div>
            <h1 className="text-lg font-semibold md:text-2xl">{table.name}</h1>
            <p className="text-sm text-muted-foreground">
              Table · {rows.length} rows · {table.columns.length} columns
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={() => setView(view === "data" ? "schema" : "data")}
          >
            {view === "data" ? "Edit schema" : "View data"}
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              setView("history");
              historyQuery.refetch();
            }}
          >
            History
          </Button>
          <Button
            onClick={addRow}
            disabled={create.isPending || view !== "data"}
          >
            <PlusIcon className="mr-2 h-4 w-4" />
            Add row
          </Button>
        </div>
      </div>
      <Separator className="my-6" />
      {view === "schema" ? (
        <SchemaEditor
          appId={appId}
          tableId={tableId}
          table={table}
          tables={tablesQuery.data?.success ? tablesQuery.data.data : []}
          onDeleted={() => router.push(`/apps/${appId}/tables`)}
        />
      ) : view === "history" ? (
        <div className="rounded-xl border bg-card shadow-sm">
          <div className="flex items-center justify-between border-b p-4">
            <h2 className="font-semibold">Audit history</h2>
            <Button variant="outline" onClick={() => setView("data")}>
              Back to data
            </Button>
          </div>
          <div className="divide-y">
            {(historyQuery.data?.success ? historyQuery.data.data : [])
              .filter((entry): entry is NonNullable<typeof entry> => !!entry)
              .map((entry) => (
                <div
                  key={entry.id}
                  className="flex items-center justify-between gap-4 p-4 text-sm"
                >
                  <div>
                    <p className="font-medium">{entry.action}</p>
                    <p className="text-muted-foreground">
                      {new Date(entry.created_at).toLocaleString()}
                      {entry.row_id ? ` · Row ${entry.row_id}` : ""}
                    </p>
                  </div>
                  {!entry.action.startsWith("undo_") && (
                    <ConfirmDialog
                      title="Undo this change?"
                      description="The table will be restored to the state before this change."
                      onConfirm={() =>
                        undo.mutate(entry.id, {
                          onSuccess: () => {
                            toast.success("Change undone");
                            historyQuery.refetch();
                            rowsQuery.refetch();
                            tableQuery.refetch();
                            setView("data");
                          },
                        })
                      }
                    >
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={undo.isPending}
                      >
                        Undo
                      </Button>
                    </ConfirmDialog>
                  )}
                </div>
              ))}
            {!historyQuery.isFetching &&
              !(
                historyQuery.data?.success && historyQuery.data.data.length
              ) && (
                <p className="p-6 text-sm text-muted-foreground">
                  No changes recorded yet.
                </p>
              )}
          </div>
        </div>
      ) : (
        <div className="rounded-xl border bg-card shadow-sm">
          <div className="flex flex-wrap items-center gap-2 border-b p-3">
            <div className="relative min-w-[220px] flex-1">
              <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Search rows..."
                className="pl-8"
              />
            </div>
            <Badge variant="secondary">{visibleRows.length} rows</Badge>
            {selected.size > 0 && (
              <BulkDeleteButton
                appId={appId}
                tableId={tableId}
                ids={selected}
                onDone={() => setSelected(new Set())}
              />
            )}
            <Button
              variant="outline"
              size="icon"
              title="Refresh"
              onClick={() => rowsQuery.refetch()}
              disabled={rowsQuery.isFetching}
            >
              <RefreshCwIcon
                className={`h-4 w-4 ${
                  rowsQuery.isFetching ? "animate-spin" : ""
                }`}
              />
            </Button>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full min-w-[760px] text-sm">
              <thead>
                <tr className="border-b bg-muted/40 text-left">
                  <th className="w-10 px-3 py-2">
                    <Checkbox
                      checked={allVisibleSelected}
                      onCheckedChange={toggleAll}
                      aria-label="Select all rows"
                    />
                  </th>
                  <th className="w-12 px-2 py-2 text-center text-xs font-medium text-muted-foreground">
                    #
                  </th>
                  {table.columns.map((column) => (
                    <SortableHeader
                      key={column.id}
                      column={column}
                      sort={sort}
                      onSort={toggleSort}
                    />
                  ))}
                  <th className="w-20 px-3 py-2" />
                </tr>
              </thead>
              <tbody>
                {visibleRows.length === 0 ? (
                  <tr>
                    <td
                      colSpan={table.columns.length + 3}
                      className="h-32 text-center text-muted-foreground"
                    >
                      {search
                        ? "No rows match your search."
                        : "This table is empty. Add a row to get started."}
                    </td>
                  </tr>
                ) : (
                  visibleRows.map((row, index) => (
                    <TableRowEditor
                      key={row.id}
                      appId={appId}
                      tableId={tableId}
                      row={row}
                      columns={table.columns}
                      index={index + 1}
                      selected={selected.has(row.id)}
                      onSelected={(checked) =>
                        setSelected((current) => {
                          const next = new Set(current);
                          if (checked) next.add(row.id);
                          else next.delete(row.id);
                          return next;
                        })
                      }
                    />
                  ))
                )}
              </tbody>
            </table>
          </div>
          <div className="flex items-center justify-between border-t px-3 py-2 text-xs text-muted-foreground">
            <span>
              Showing {visibleRows.length} of {rows.length} rows
            </span>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="icon" disabled>
                <ChevronLeftIcon className="h-4 w-4" />
              </Button>
              <span>Page 1</span>
              <Button variant="ghost" size="icon" disabled>
                <ChevronLeftIcon className="h-4 w-4 rotate-180" />
              </Button>
            </div>
          </div>
        </div>
      )}
    </AppLayout>
  );
}

function SchemaEditor({
  appId,
  tableId,
  table,
  tables,
  onDeleted,
}: {
  appId: string;
  tableId: string;
  table: Table;
  tables: (Table | undefined)[];
  onDeleted: () => void;
}) {
  const update = useTableUpdateMutation(appId, tableId);
  const remove = useTableDeleteMutation(appId, tableId);
  const [name, setName] = useState(table.name);
  const [columns, setColumns] = useState(table.columns);
  const [confirmSchemaSave, setConfirmSchemaSave] = useState(false);
  const edit = (index: number, patch: Partial<TableColumn>) =>
    setColumns((current) =>
      current.map((column, i) =>
        i === index ? { ...column, ...patch } : column
      )
    );
  const save = () => {
    const changedIds = table.columns.some(
      (column) => !columns.some((next) => next.id === column.id)
    );
    if (changedIds) {
      setConfirmSchemaSave(true);
      return;
    }
    saveSchema();
  };
  const saveSchema = () =>
    update.mutate(
      { name, columns },
      {
        onSuccess: (res) =>
          res.success
            ? toast.success("Schema saved")
            : toast.error(res.error.message),
      }
    );
  const deleteTable = () =>
    remove.mutate(undefined, {
      onSuccess: (res) =>
        res.success ? onDeleted() : toast.error(res.error.message),
    });
  return (
    <div className="max-w-5xl rounded-xl border bg-card shadow-sm">
      <div className="flex items-center justify-between border-b p-4">
        <div>
          <h2 className="font-semibold">Table schema</h2>
          <p className="text-sm text-muted-foreground">
            Configure columns, required fields, and references.
          </p>
        </div>
        <Button onClick={save} disabled={update.isPending}>
          Save schema
        </Button>
      </div>
      <div className="space-y-5 p-4">
        <div className="max-w-md">
          <label className="text-sm font-medium">Table name</label>
          <Input
            className="mt-1"
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </div>
        <div className="overflow-x-auto rounded-md border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/40 text-left">
                <th className="px-3 py-2">Column</th>
                <th className="px-3 py-2">Type</th>
                <th className="px-3 py-2">Required</th>
                <th className="px-3 py-2">Reference</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {columns.map((column, index) => (
                <tr className="border-b last:border-0" key={index}>
                  <td className="p-2">
                    <Input
                      value={column.name}
                      onChange={(event) =>
                        edit(index, {
                          name: event.target.value,
                          id: event.target.value
                            .toLowerCase()
                            .replace(/[^a-z0-9_]/g, "_"),
                        })
                      }
                    />
                  </td>
                  <td className="p-2">
                    <CustomSelect
                      value={column.type}
                      options={[
                        "text",
                        "number",
                        "boolean",
                        "date",
                        "json",
                        "reference",
                      ].map((type) => ({ value: type, label: type }))}
                      placeholder="Type"
                      searchable={false}
                      className="h-9 w-36"
                      onChange={(value) =>
                        edit(index, {
                          type: value,
                          reference_table_id: undefined,
                        })
                      }
                    />
                  </td>
                  <td className="p-2">
                    <Checkbox
                      checked={column.required ?? false}
                      onCheckedChange={(checked) =>
                        edit(index, { required: checked === true })
                      }
                    />
                  </td>
                  <td className="p-2">
                    {column.type === "reference" ? (
                      <CustomSelect
                        value={column.reference_table_id ?? ""}
                        options={tables
                          .filter((item) => item && item.id !== tableId)
                          .filter((item): item is Table => !!item)
                          .map((item) => ({
                            value: item.id,
                            label: item.name,
                          }))}
                        placeholder="Select table"
                        searchable={false}
                        className="min-w-48"
                        onChange={(value) =>
                          edit(index, { reference_table_id: value })
                        }
                      />
                    ) : (
                      <span className="text-xs text-muted-foreground">
                        Not applicable
                      </span>
                    )}
                  </td>
                  <td className="p-2 text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() =>
                        setColumns((current) =>
                          current.filter((_, i) => i !== index)
                        )
                      }
                    >
                      Remove
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <Button
          variant="outline"
          onClick={() =>
            setColumns((current) => [
              ...current,
              {
                id: `column_${current.length + 1}`,
                name: `Column ${current.length + 1}`,
                type: "text",
              },
            ])
          }
        >
          <PlusIcon className="mr-2 h-4 w-4" />
          Add column
        </Button>
        <div className="flex justify-end border-t pt-4">
          <ConfirmDialog
            title={`Delete "${table.name}"?`}
            description="This permanently deletes the table and all of its rows."
            onConfirm={deleteTable}
          >
            <Button variant="destructive" disabled={remove.isPending}>
              <Trash2Icon className="mr-2 h-4 w-4" />
              Delete table
            </Button>
          </ConfirmDialog>
        </div>
      </div>
      <AlertDialog open={confirmSchemaSave} onOpenChange={setConfirmSchemaSave}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Change table columns?</AlertDialogTitle>
            <AlertDialogDescription>
              Renaming or removing columns can make existing row data
              inaccessible.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={saveSchema}>Continue</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function SortableHeader({
  column,
  sort,
  onSort,
}: {
  column: TableColumn;
  sort: Sort | null;
  onSort: (id: string) => void;
}) {
  const active = sort?.column === column.id;
  return (
    <th className="whitespace-nowrap px-3 py-2 font-medium">
      <button
        type="button"
        className="inline-flex items-center gap-1.5 hover:text-primary"
        onClick={() => onSort(column.id)}
      >
        {column.name}
        <span className="text-[10px] text-muted-foreground">{column.type}</span>
        {active ? (
          sort.direction === "asc" ? (
            <ArrowUpIcon className="h-3 w-3" />
          ) : (
            <ArrowDownIcon className="h-3 w-3" />
          )
        ) : (
          <span className="text-muted-foreground/50">↕</span>
        )}
      </button>
    </th>
  );
}

function TableRowEditor({
  appId,
  tableId,
  row,
  columns,
  index,
  selected,
  onSelected,
}: {
  appId: string;
  tableId: string;
  row: TableRow;
  columns: TableColumn[];
  index: number;
  selected: boolean;
  onSelected: (checked: boolean) => void;
}) {
  const update = useTableRowUpdateMutation(appId, tableId, row.id);
  const remove = useTableRowDeleteMutation(appId, tableId);
  const [data, setData] = useState(row.data);
  const save = (column: TableColumn, value: string) => {
    const typed =
      value === ""
        ? value
        : column.type === "number"
        ? Number(value)
        : column.type === "boolean"
        ? value === "true"
        : value;
    const next = { ...data, [column.id]: typed };
    setData(next);
    update.mutate({ data: next });
  };
  return (
    <tr
      className={`border-b last:border-0 ${
        selected ? "bg-primary/5" : "hover:bg-muted/30"
      }`}
    >
      <td className="px-3 py-1.5">
        <Checkbox
          checked={selected}
          onCheckedChange={(checked) => onSelected(checked === true)}
          aria-label={`Select row ${index}`}
        />
      </td>
      <td className="px-2 py-1.5 text-center font-mono text-xs text-muted-foreground">
        {index}
      </td>
      {columns.map((column) => (
        <td className="min-w-[150px] max-w-[320px] p-1.5" key={column.id}>
          <CellEditor
            column={column}
            value={data[column.id]}
            onSave={(value) => save(column, value)}
          />
        </td>
      ))}
      <td className="px-2 py-1.5 text-right">
        <ConfirmDialog
          title="Delete this row?"
          description="This permanently deletes the row and its data."
          onConfirm={() => remove.mutate(row.id)}
        >
          <Button
            variant="ghost"
            size="icon"
            className="text-muted-foreground hover:text-destructive"
            title={`Delete row ${row.id}`}
            disabled={remove.isPending}
          >
            <Trash2Icon className="h-4 w-4" />
          </Button>
        </ConfirmDialog>
      </td>
    </tr>
  );
}

function CellEditor({
  column,
  value,
  onSave,
}: {
  column: TableColumn;
  value: unknown;
  onSave: (value: string) => void;
}) {
  const text =
    value == null
      ? ""
      : typeof value === "object"
      ? JSON.stringify(value)
      : String(value);
  if (column.type === "boolean")
    return (
      <CustomSelect
        value={text}
        options={[
          { value: "", label: "Empty" },
          { value: "true", label: "True" },
          { value: "false", label: "False" },
        ]}
        placeholder="Select..."
        searchable={false}
        className="h-8 w-full border-transparent bg-transparent px-2 hover:border-input"
        onChange={onSave}
      />
    );
  if (column.type === "json")
    return (
      <Textarea
        defaultValue={text}
        onBlur={(event) => onSave(event.target.value)}
        className="min-h-8 resize-none border-transparent font-mono text-xs hover:border-input focus:border-input"
        rows={1}
      />
    );
  return (
    <Input
      type={column.type === "date" ? "datetime-local" : "text"}
      inputMode={column.type === "number" ? "decimal" : undefined}
      defaultValue={text}
      onBlur={(event) => onSave(event.target.value)}
      className="h-8 border-transparent bg-transparent hover:border-input focus:border-input"
    />
  );
}

function BulkDeleteButton({
  appId,
  tableId,
  ids,
  onDone,
}: {
  appId: string;
  tableId: string;
  ids: Set<string>;
  onDone: () => void;
}) {
  const remove = useTableRowDeleteMutation(appId, tableId);
  const deleteRows = async () => {
    for (const id of ids) await remove.mutateAsync(id);
    onDone();
  };
  return (
    <ConfirmDialog
      title={`Delete ${ids.size} rows?`}
      description="This permanently deletes the selected rows and their data."
      onConfirm={deleteRows}
    >
      <Button variant="destructive" size="sm" disabled={remove.isPending}>
        <Trash2Icon className="mr-2 h-4 w-4" />
        Delete {ids.size}
      </Button>
    </ConfirmDialog>
  );
}

TablePage.getLayout = getAppShellLayout;
