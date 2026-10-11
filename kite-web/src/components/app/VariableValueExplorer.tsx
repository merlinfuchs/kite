import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "../ui/input";
import { Badge } from "../ui/badge";
import { Skeleton } from "../ui/skeleton";
import { useEffect, useState } from "react";
import { RefreshCwIcon } from "lucide-react";
import { useVariable } from "@/lib/hooks/api";
import { useVariableValuesQuery } from "@/lib/api/queries";
import { useVariableValueDeleteMutation } from "@/lib/api/mutations";
import { useAppId, useVariableId } from "@/lib/hooks/params";
import { Variable, VariableValue } from "@/lib/types/wire.gen";
import { cn, formatDateTime, formatNumber } from "@/lib/utils";
import { toast } from "sonner";
import ConfirmDialog from "../common/ConfirmDialog";
import VariableValueDialog, {
  isValueReadOnly,
  variableValueTypes,
} from "./VariableValueDialog";

const pageSize = 25;

export default function VariableValueExplorer() {
  const variable = useVariable();

  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(0);

  // Debounced so not every keystroke is a request.
  useEffect(() => {
    const timeout = setTimeout(() => {
      setSearch(searchInput.trim());
      setPage(0);
    }, 300);
    return () => clearTimeout(timeout);
  }, [searchInput]);

  const query = useVariableValuesQuery(useAppId(), useVariableId(), {
    search,
    limit: pageSize,
    offset: page * pageSize,
  });

  const data = query.data?.success ? query.data.data : undefined;
  const error = query.data && !query.data.success ? query.data.error : null;
  const values = data?.values ?? [];
  const total = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));

  // Deleting the last value of the last page leaves the page empty.
  useEffect(() => {
    if (data && page > 0 && page >= pageCount) {
      setPage(pageCount - 1);
    }
  }, [data, page, pageCount]);

  // An unscoped variable has one value, so there is nothing to search or add
  // once it's set.
  const scoped = !!variable?.scoped;
  const canAdd =
    !!variable &&
    (scoped || (!!data && !values.some((v) => v!.scope === null)));
  // Flows can store scoped values in an unscoped variable too.
  const showScope = scoped || values.some((v) => v!.scope !== null);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{scoped ? "Stored Values" : "Stored Value"}</CardTitle>
        <CardDescription>
          {scoped
            ? "View and edit the values your flows have stored in this variable, one for each scope."
            : "View and edit the value your flows have stored in this variable."}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center space-x-3">
          {scoped && (
            <Input
              placeholder="Search by scope..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="max-w-sm"
            />
          )}
          <div className="flex-1" />
          <Button
            variant="outline"
            size="icon"
            onClick={() => query.refetch()}
            disabled={query.isFetching}
            title="Refresh"
          >
            <RefreshCwIcon
              className={cn("h-4 w-4", query.isFetching && "animate-spin")}
            />
          </Button>
          {canAdd && (
            <VariableValueDialog variable={variable}>
              <Button>{scoped ? "Add value" : "Set value"}</Button>
            </VariableValueDialog>
          )}
        </div>

        {!variable || query.isPending ? (
          <Skeleton className="h-32" />
        ) : error ? (
          <div className="rounded-md border px-4 py-8 text-center text-sm text-muted-foreground">
            Failed to load values: {error.message} ({error.code})
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  {showScope && <TableHead>Scope</TableHead>}
                  <TableHead>Value</TableHead>
                  <TableHead className="hidden sm:table-cell">Type</TableHead>
                  <TableHead className="hidden sm:table-cell">
                    Updated
                  </TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {values.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={showScope ? 5 : 4}
                      className="h-24 text-center text-muted-foreground"
                    >
                      {search
                        ? "No values match your search."
                        : "No values have been stored yet."}
                    </TableCell>
                  </TableRow>
                ) : (
                  values.map((value) => (
                    <VariableValueRow
                      variable={variable}
                      value={value!}
                      showScope={showScope}
                      key={value!.scope ?? ""}
                    />
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
      {data && (scoped ? total > 0 : total > 1) && (
        <CardFooter className="flex flex-wrap items-center justify-between gap-3 border-t px-6 py-4">
          <div className="text-sm text-muted-foreground">
            {formatNumber(page * pageSize + 1)} -{" "}
            {formatNumber(page * pageSize + values.length)} of{" "}
            {formatNumber(total)} values
          </div>
          <div className="flex space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p + 1)}
              disabled={page + 1 >= pageCount}
            >
              Next
            </Button>
          </div>
        </CardFooter>
      )}
    </Card>
  );
}

function VariableValueRow({
  variable,
  value,
  showScope,
}: {
  variable: Variable;
  value: VariableValue;
  showScope: boolean;
}) {
  const deleteMutation = useVariableValueDeleteMutation(
    useAppId(),
    variable.id
  );

  function remove() {
    deleteMutation.mutate(value.scope, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Value deleted!");
        } else {
          toast.error(
            `Failed to delete value: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  const readOnly = isValueReadOnly(variable, value);
  const typeLabel =
    variableValueTypes.find((t) => t.value === value.type)?.label ?? value.type;

  return (
    <TableRow>
      {showScope && (
        <TableCell className="font-mono text-xs max-w-[12rem] truncate">
          {value.scope ?? <span className="text-muted-foreground">none</span>}
        </TableCell>
      )}
      <TableCell className="font-mono text-xs max-w-[12rem] md:max-w-xs truncate">
        {value.value === "" ? (
          <span className="text-muted-foreground">empty</span>
        ) : (
          // Enough for the cell, long values are read in the dialog.
          value.value.slice(0, 200)
        )}
      </TableCell>
      <TableCell className="hidden sm:table-cell">
        <Badge variant="secondary" className="whitespace-nowrap">
          {typeLabel}
        </Badge>
      </TableCell>
      <TableCell className="hidden sm:table-cell text-muted-foreground whitespace-nowrap">
        {formatDateTime(new Date(value.updated_at))}
      </TableCell>
      <TableCell>
        <div className="flex justify-end space-x-2">
          <VariableValueDialog variable={variable} value={value}>
            <Button size="sm" variant="outline">
              {readOnly ? "View" : "Edit"}
            </Button>
          </VariableValueDialog>
          <ConfirmDialog
            title="Are you sure that you want to delete this value?"
            description="Flows will find the variable empty for this scope. This can't be undone."
            onConfirm={remove}
          >
            <Button size="sm" variant="ghost">
              Delete
            </Button>
          </ConfirmDialog>
        </div>
      </TableCell>
    </TableRow>
  );
}
