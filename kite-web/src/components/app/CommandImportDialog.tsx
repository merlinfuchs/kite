import { ReactNode, useCallback, useState } from "react";
import { useRouter } from "next/router";
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

  const router = useRouter();
  const appId = useAppId();
  const commandsImportMutation = useCommandsImportMutation(appId);

  const handleImport = useCallback(() => {
    const trimmed = shareCode.trim();

    if (!trimmed) {
      toast.error("Paste code before importing");
      return;
    }

    let parsed: unknown;
    try {
      parsed = JSON.parse(trimmed);
    } catch {
      toast.error(
        "Error importing command: invalid share code"
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
        "Error importing command: invalid share code"
      );
      return;
    }

    const command = parsed as CommandCreateRequest;

    commandsImportMutation.mutate(
      { commands: [command] },
      {
        onSuccess: (res) => {
          if (res.success) {
            const imported = res.data[0];

            if (!imported) {
              toast.error(
                `Error importing command: command not found`
              );
              setShareCode("");
              setDialogOpen(false);
              return;
            }

            toast.success("Command imported!");
            setShareCode("");
            setDialogOpen(false);

            setTimeout(
              () =>
                router.push({
                  pathname: "/apps/[appId]/commands/[cmdId]",
                  query: { appId, cmdId: imported.id },
                }),
              500
            );
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
            Import a command using a share code.
          </DialogDescription>
        </DialogHeader>
        <div className="mb-3 mt-2 space-y-3">
          <div>
            <div className="font-medium mb-0.5">Share Code</div>
            <div className="text-sm text-muted-foreground mb-3"></div>
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