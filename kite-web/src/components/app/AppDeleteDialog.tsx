import { useAppDeleteMutation } from "@/lib/api/mutations";
import { App } from "@/lib/types/wire.gen";
import { useRouter } from "next/router";
import { FormEvent, ReactNode, useState } from "react";
import { toast } from "sonner";
import { track } from "@/lib/analytics";
import LoadingButton from "../common/LoadingButton";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { Label } from "../ui/label";

export default function AppDeleteDialog({
  app,
  children,
}: {
  app: App;
  children: ReactNode;
}) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");

  const deleteMutation = useAppDeleteMutation(app.id);

  // Typing the name makes sure the right app is deleted.
  const confirmed = name.trim() === app.name.trim();

  function onOpenChange(open: boolean) {
    setOpen(open);
    if (!open) setName("");
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!confirmed) return;

    deleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          track("app_deleted");
          toast.success(`${app.name} deleted!`);
          onOpenChange(false);
          router.push("/apps");
        } else {
          toast.error(
            `Failed to delete app: ${res.error.message} (${res.error.code})`
          );
        }
      },
      onError(err) {
        toast.error(`Failed to delete app: ${err.message}`);
      },
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <form onSubmit={onSubmit} className="space-y-5">
          <DialogHeader>
            <DialogTitle className="break-words">
              Do you want to delete {app.name}?
            </DialogTitle>
            <DialogDescription>
              This deletes {app.name} and all of its commands, event listeners,
              message templates, variables and logs from Kite. It can&apos;t be
              undone. Your bot stays in the Discord Developer Portal.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="delete-app-name">
              Type <span className="font-semibold">{app.name}</span> to confirm
            </Label>
            <Input
              id="delete-app-name"
              autoComplete="off"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <DialogFooter className="gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            {/* LoadingButton only disables itself while loading. */}
            {confirmed ? (
              <LoadingButton
                type="submit"
                variant="destructive"
                loading={deleteMutation.isPending}
              >
                Delete app
              </LoadingButton>
            ) : (
              <Button type="submit" variant="destructive" disabled>
                Delete app
              </Button>
            )}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
