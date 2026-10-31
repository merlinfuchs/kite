import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Separator } from "../ui/separator";
import ConfirmDialog from "../common/ConfirmDialog";
import { useApp, useUser } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { useAppCollaboratorDeleteMutation } from "@/lib/api/mutations";
import { useRouter } from "next/router";
import { toast } from "sonner";
import AppDeleteDialog from "./AppDeleteDialog";
import AppDuplicateDialog from "./AppDuplicateDialog";

export default function AppSettingsDangerZone() {
  const router = useRouter();
  const appId = useAppId();
  const app = useApp();
  const user = useUser();
  const leaveMutation = useAppCollaboratorDeleteMutation(appId);

  if (!app || !user) return null;
  const userId = user.id;
  const isOwner = app.owner_user_id === userId;

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

  if (!isOwner) {
    return (
      <Card>
        <CardContent className="pt-6">
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
              onConfirm={leave}
            >
              <Button variant="destructive" className="flex-none">
                Leave app
              </Button>
            </ConfirmDialog>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-destructive/50">
      <CardHeader>
        <CardTitle className="text-destructive">Danger Zone</CardTitle>
        <CardDescription>
          These change your whole app, so make sure you picked the right one.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Duplicate App</div>
            <div className="text-muted-foreground">
              Copy everything you built in this app to another bot. This app
              stays as it is.
            </div>
          </div>
          <AppDuplicateDialog app={app}>
            <Button variant="outline" className="flex-none">
              Duplicate app
            </Button>
          </AppDuplicateDialog>
        </div>
        <Separator />
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-bold pb-1">Delete App</div>
            <div className="text-muted-foreground">
              Deleting your app will remove it from Kite and delete all
              associated data.
            </div>
          </div>
          <AppDeleteDialog app={app}>
            <Button
              variant="destructive"
              className="space-x-2 flex items-center flex-none"
            >
              <div>Delete app</div>
            </Button>
          </AppDeleteDialog>
        </div>
      </CardContent>
    </Card>
  );
}
