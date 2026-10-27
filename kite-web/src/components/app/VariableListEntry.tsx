import { StretchHorizontalIcon, VariableIcon } from "lucide-react";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import BundleInstallBadge from "./BundleInstallBadge";
import { Button } from "../ui/button";
import { useRouter } from "next/router";
import Link from "next/link";
import { Variable } from "@/lib/types/wire.gen";
import ConfirmDialog from "../common/ConfirmDialog";
import { useVariableDeleteMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";

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
            `Failed to delete variable: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Card className="relative">
      <div className="px-6 pt-4 sm:absolute sm:top-0 sm:right-0 sm:pt-3 sm:pr-4">
        <div className="flex items-center space-x-2">
          <StretchHorizontalIcon className="h-5 w-5 text-muted-foreground" />
          <div className="text-sm">{variable.total_values || 0} values</div>
        </div>
      </div>
      <CardHeader className="pt-3 sm:pt-6 sm:pr-48">
        <CardTitle className="text-base flex flex-wrap items-center gap-x-2 gap-y-1">
          <div className="flex items-center gap-2 min-w-0">
            <VariableIcon className="h-5 w-5 flex-none text-muted-foreground" />
            <div className="min-w-0 break-words">{variable.name}</div>
          </div>
          <BundleInstallBadge bundleInstallId={variable.bundle_install_id} />
        </CardTitle>
        <CardDescription className="text-sm">
          This variable stores{" "}
          {variable.scoped
            ? "multiple values scoped by a specific key."
            : "one value."}
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
