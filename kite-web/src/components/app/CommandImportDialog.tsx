import { ReactNode, useCallback, useState } from "react";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { useCommandsImportMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { CommandCreateRequest } from "@/lib/types/wire.gen";

export function CommandImportDialog({ children }: { children: ReactNode }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [shareCode, setShareCode] = useState("");

  const appId = useAppId();
  const commandsImportMutation = useCommandsImportMutation(appId);

  const handleImport = useCallback(() => {
    const trimmed = shareCode.trim();

    if (!trimmed) {
      toast.error("Please paste a share code before importing");
      return;
    }

    let parsed: unknown;
    try {
      parsed = JSON.parse(trimmed);
    } catch {
      toast.error(
        "Invalid share code... please check you copied it correctly"
      );
      return;
    }

    if (
      typeof parsed !== "object" ||
      parsed === null ||
      !("flow_source" in parsed) ||
      !("enabled" in parsed)
    ) {
      toast.error(
        "This doesn't look like a valid command share code, please check you copied it correctly"
      );
      return;
    }

    const command = parsed as CommandCreateRequest;

    commandsImportMutation.mutate(
      { commands: [command] },
      {
        onSuccess: (res) => {
          if (res.success) {
            toast.success("Command imported successfully");
            setShareCode("");
            setDialogOpen(false);
          } else {
            toast.error(
              `Failed to import command: ${res.error.message} (${res.error.code})`
            );
          }
        },
        onError: () => {
          toast.error(
            "Failed to import command, please check the share code and try again"
          );
        },
      }
    );
  }, [shareCode, commandsImportMutation]);

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(open) => {
        setDialogOpen(open);
        if (!open) setShareCode("");
      }}
    >
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="mb-1">Import Command</DialogTitle>
          <DialogDescription>
            Import a command using a generated share code.
          </DialogDescription>
        </DialogHeader>
        <div className="mb-3 mt-2 space-y-3">
          <div>
            <div className="font-medium mb-0.5">Share Code</div>
            <div className="text-sm text-muted-foreground mb-3">
            </div>
            <Input
              value={shareCode}
              onChange={(e) => setShareCode(e.target.value)}
              placeholder=""
            />
          </div>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button
            onClick={handleImport}
            disabled={commandsImportMutation.isPending}
          >
            {commandsImportMutation.isPending ? "Importing..." : "Import"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}