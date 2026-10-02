import { useAppDeleteMutation } from "@/lib/api/mutations";
import ConfirmDialog from "../common/ConfirmDialog";
import { Button } from "../ui/button";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { useRouter } from "next/router";
import { Card, CardContent } from "../ui/card";

export default function AppSettingsDelete() {
  const router = useRouter();
  const deleteMutation = useAppDeleteMutation(useAppId());

  function remove() {
    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
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
