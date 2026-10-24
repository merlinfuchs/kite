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
import { useQueryClient } from "@tanstack/react-query";

export default function AppSettingsDelete() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const appId = useAppId();
  const app = useApp();
  const user = useUser();
  const deleteMutation = useAppDeleteMutation(appId);
  const leaveMutation = useAppCollaboratorDeleteMutation(appId);

  if (app && user && app.owner_user_id !== user.id) {
    return (
      <Card>
        <CardContent className="pt-6 space-y-5">
          <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div className="font-bold pb-1">Leave App</div>
              <div className="text-muted-foreground">
                You won&apos;t be able to manage this app anymore until the
                owner adds you again.
              </div>
            </div>
            <ConfirmDialog
              title="Are you sure that you want to leave this app?"
              description="The app keeps running, but you lose access to it."
              onConfirm={() =>
                leaveMutation.mutate(user.id, {
                  onSuccess(res) {
                    if (res.success) {
                      toast.success("Left app!");
                      queryClient.invalidateQueries({ queryKey: ["apps"] });
                      router.push("/apps");
                    } else {
                      toast.error(
                        `Failed to leave app: ${res.error.message} (${res.error.code})`
                      );
                    }
                  },
                })
              }
            >
              <Button variant="destructive">Leave app</Button>
            </ConfirmDialog>
          </div>
        </CardContent>
      </Card>
    );
  }

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

  return (
    <Card>
      <CardContent className="pt-6 space-y-5">
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Delete App</div>
            <div className="text-muted-foreground">
              Deleting your app will remove it from Kite and delete all
              associated data.
            </div>
          </div>
          <ConfirmDialog
            title="Are you sure that you want to delete this app?"
            description="This removes all associated data and cannot be undone."
            onConfirm={remove}
          >
            <Button
              variant="destructive"
              className="space-x-2 flex items-center"
            >
              <div>Delete app</div>
            </Button>
          </ConfirmDialog>
        </div>
      </CardContent>
    </Card>
  );
}
