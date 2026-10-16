import {
  CheckIcon,
  CopyPlusIcon,
  EllipsisIcon,
  Share2Icon,
  SquareFunctionIcon,
  Trash2Icon,
} from "lucide-react";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Button } from "../ui/button";
import { useRouter } from "next/router";
import Link from "next/link";
import { FlowFunction } from "@/lib/types/wire.gen";
import ConfirmDialog from "../common/ConfirmDialog";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { useFunctionDeleteMutation } from "@/lib/api/mutations";
import { formatDateTime } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import {
  FunctionUsage,
  describeFunctionUsages,
  functionToShared,
  getFunctionParameters,
} from "@/lib/flow/functions";
import FunctionDuplicateDialog from "./FunctionDuplicateDialog";
import FlowExportDialog from "./FlowExportDialog";
import { Badge } from "../ui/badge";

export default function FunctionListEntry({
  function: fn,
  usages,
}: {
  function: FlowFunction;
  // Undefined while they're loading.
  usages?: FunctionUsage[];
}) {
  const router = useRouter();

  const deleteMutation = useFunctionDeleteMutation(useAppId(), fn.id);
  const parameters = getFunctionParameters(fn.flow_source);

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Function deleted!");
        } else {
          toast.error(
            `Failed to delete function: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Card>
      <div className="float-right pt-3 pr-4">
        <div className="flex items-center space-x-2">
          <CheckIcon className="h-5 w-5 text-green-500" />
          <div className="text-sm text-muted-foreground">
            {formatDateTime(new Date(fn.updated_at))}
          </div>
        </div>
      </div>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <SquareFunctionIcon className="h-5 w-5 text-muted-foreground" />
          <div>{fn.name}</div>
        </CardTitle>
        <CardDescription className="text-sm">{fn.description}</CardDescription>
        {usages && (
          <div
            className="text-xs text-muted-foreground"
            title={usages.map((u) => u.label).join("\n")}
          >
            {usages.length > 0
              ? `Used by ${describeFunctionUsages(usages, 3)}`
              : "Not used by any command, event listener or message template yet"}
          </div>
        )}
        {parameters.length > 0 && (
          <div className="flex flex-wrap gap-2 pt-1">
            {parameters.map((p) => (
              <Badge variant="secondary" key={p.name}>
                {p.name}
              </Badge>
            ))}
          </div>
        )}
      </CardHeader>
      <CardFooter className="flex space-x-3">
        <Button size="sm" variant="outline" asChild>
          <Link
            href={{
              pathname: "/apps/[appId]/functions/[functionId]",
              query: {
                appId: router.query.appId,
                functionId: fn.id,
              },
            }}
          >
            Manage
          </Link>
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button size="icon" variant="ghost">
              <EllipsisIcon className="h-5 w-5 text-muted-foreground" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuGroup>
              <ConfirmDialog
                title="Are you sure that you want to delete this function?"
                description={
                  usages && usages.length > 0
                    ? `This function is used by ${describeFunctionUsages(
                        usages
                      )}. Their Call function blocks will fail until you pick another function. This cannot be undone.`
                    : "This will remove the function from your app and cannot be undone."
                }
                onConfirm={remove}
              >
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Trash2Icon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Delete Function
                </DropdownMenuItem>
              </ConfirmDialog>
              <FunctionDuplicateDialog function={fn}>
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <CopyPlusIcon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Duplicate Function
                </DropdownMenuItem>
              </FunctionDuplicateDialog>
              <FlowExportDialog
                title="Export Function"
                type="function"
                shareData={{ function: functionToShared(fn) }}
                flows={[fn.flow_source]}
                excludeFunctionIds={[fn.id]}
              >
                <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                  <Share2Icon className="h-4 w-4 mr-2 text-muted-foreground" />
                  Export Function
                </DropdownMenuItem>
              </FlowExportDialog>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardFooter>
    </Card>
  );
}
