import { ReactNode, useEffect, useState } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogHeader,
  DialogTrigger,
  DialogFooter,
  DialogClose,
  DialogDescription,
} from "../ui/dialog";
import { useCommandsDeployMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import LoadingButton from "../common/LoadingButton";
import { Button } from "../ui/button";
import { toast } from "sonner";
import { useEdges } from "@xyflow/react";

export function CommandDeployDialog({
  children,
  open,
  onOpenChange,
}: {
  children?: ReactNode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const appId = useAppId();
  const deployMutation = useCommandsDeployMutation(appId);

  const [error, setError] = useState<string | null>(null);

  function onDeploy() {
    deployMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          if (res.data.deployed) {
            toast.success(
              "Commands deployed successfully! Restart your Discord client if the changes are not visible immediately."
            );
            onOpenChange(false);
            setError(null);
          } else {
            setError(JSON.stringify(res.data.error, null, 2));
          }
        } else {
          toast.error(
            `Failed to deploy commands: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  useEffect(() => {
    if (open) {
      setError(null);
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {children && <DialogTrigger asChild>{children}</DialogTrigger>}
      <DialogContent className="sm:max-w-xl w-full overflow-x-auto">
        <DialogHeader>
          <DialogTitle>Deploy Commands</DialogTitle>
          <DialogDescription>
            Deploy the commands so they are available inside Discord. Whenever
            you make changes to a command&apos;s appearance, you need to deploy
            it again for the changes to take effect.
            <br />
            <br />
            You can only deploy commands once every 30 seconds, so only deploy
            when you have actually made changes.
          </DialogDescription>
        </DialogHeader>
        {error && (
          <div>
            <div className="text-destructive font-medium mb-2">
              Deployment Error
            </div>
            <div className="bg-destructive/10 border border-destructive p-4 rounded-md font-mono whitespace-pre text-sm overflow-auto w-full">
              {error}
            </div>
          </div>
        )}
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <LoadingButton loading={deployMutation.isPending} onClick={onDeploy}>
            Deploy commands
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
