import {
  useAppCollaboratorDeleteMutation,
  useAppDeleteMutation,
} from "@/lib/api/mutations";
import { useApp, useUser } from "@/lib/hooks/api";
import ConfirmDialog from "../common/ConfirmDialog";
import { Button } from "../ui/button";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { useRouter } from "next/router";
import { Card, CardContent } from "../ui/card";
import { track } from "@/lib/analytics";

export default function AppSettingsDelete() {
  const router = useRouter();
  const appId = useAppId();
  const app = useApp();
  const user = useUser();
  const deleteMutation = useAppDeleteMutation(appId);
  const leaveMutation = useAppCollaboratorDeleteMutation(appId);

  if (!app || !user) return null;
  const userId = user.id;
  const isOwner = app.owner_user_id === userId;

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          track("app_deleted");
          toast.success("App deleted!");
          router.push("/apps");
        } else {
          toast.error(
            `Failed to delete app: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  function leave() {
    leaveMutation.mutate(userId, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Left app!");
          router.push("/apps");
        } else {
          toast.error(
            `Failed to leave app: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <Card>
      <CardContent className="pt-6 space-y-5">
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">
              {isOwner ? "Delete App" : "Leave App"}
            </div>
            <div className="text-muted-foreground">
              {isOwner
                ? "Deleting your app will remove it from Kite and delete all associated data."
                : "You won't be able to manage this app anymore until the owner adds you again."}
            </div>
          </div>
          <ConfirmDialog
            title={
              isOwner
                ? "Are you sure that you want to delete this app?"
                : "Are you sure that you want to leave this app?"
            }
            description={
              isOwner
                ? "This removes all associated data and cannot be undone."
                : "The app keeps running, but you lose access to it."
            }
            onConfirm={isOwner ? remove : leave}
          >
            <Button
              variant="destructive"
              className="space-x-2 flex items-center"
            >
              <div>{isOwner ? "Delete app" : "Leave app"}</div>
            </Button>
          </ConfirmDialog>
        </div>
      </CardContent>
    </Card>
  );
}
