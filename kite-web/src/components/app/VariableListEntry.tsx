import { CheckIcon, SlashSquareIcon, VariableIcon } from "lucide-react";
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
import { Variable } from "@/lib/types/wire.gen";
import ConfirmDialog from "../common/ConfirmDialog";
import { useVariableDeleteMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { formatDateTime } from "@/lib/utils";

export default function VariableListEntry({
  variable,
}: {
  variable: Variable;
}) {
  const router = useRouter();

  const deleteMutation = useVariableDeleteMutation(useAppId(), variable.id);

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Variable deleted!");
        } else {
          toast.error(
            `Failed to load app: ${res.error.message} (${res.error.code})`
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
            {formatDateTime(new Date(variable.updated_at))}
          </div>
        </div>
      </div>
      <CardHeader>
        <CardTitle className="text-base flex items-center space-x-2">
          <VariableIcon className="h-5 w-5 text-muted-foreground" />
          <div>{variable.name}</div>
        </CardTitle>
        <CardDescription className="text-sm">
          {variable.type} / {variable.scope}
        </CardDescription>
      </CardHeader>
      <CardFooter className="flex space-x-3">
        <Button size="sm" variant="outline" asChild>
          <Link
            href={{
              pathname: "/apps/[appId]/variables/[variableId]",
              query: {
                appId: router.query.appId,
                variableId: variable.id,
              },
            }}
          >
            Manage
          </Link>
        </Button>
        <ConfirmDialog
          title="Are you sure that you want to delete this variable?"
          description="This will remove the variable from your app and cannot be undone."
          onConfirm={remove}
        >
          <Button
            size="sm"
            variant="ghost"
            className="space-x-2 flex items-center"
          >
            <div>Delete</div>
          </Button>
        </ConfirmDialog>
      </CardFooter>
    </Card>
  );
}
