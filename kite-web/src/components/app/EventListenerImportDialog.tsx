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
import { useEventListenersImportMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { EventListenerCreateRequest } from "@/lib/types/wire.gen";

export function EventListenerImportDialog({ children }: { children: ReactNode }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [shareCode, setShareCode] = useState("");

  const router = useRouter();
  const appId = useAppId();
  const eventListenersImportMutation = useEventListenersImportMutation(appId);

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
        "Error importing event listener: invalid share code"
      );
      return;
    }

    if (
      typeof parsed !== "object" ||
      parsed === null ||
      !("flow_source" in parsed) ||
      !("enabled" in parsed) ||
      !("source" in parsed) ||
      !Array.isArray((parsed as any).flow_source?.nodes) ||
      !(parsed as any).flow_source.nodes.some(
        (node: any) => node?.type === "entry_event"
      )
    ) {
      toast.error(
        "Error importing event listener: invalid share code"
      );
      return;
    }

    const listener = parsed as EventListenerCreateRequest;

    eventListenersImportMutation.mutate(
      { event_listeners: [listener] },
      {
        onSuccess: (res) => {
          if (res.success) {
            const imported = res.data[0];

            if (!imported) {
              toast.error(
                `Error importing event listener: event listener not found`
              );
              setShareCode("");
              setDialogOpen(false);
              return;
            }

            toast.success("Event listener imported!");
            setShareCode("");
            setDialogOpen(false);

            setTimeout(
              () =>
                router.push({
                  pathname: "/apps/[appId]/events/[eventId]",
                  query: { appId, eventId: imported.id },
                }),
              500
            );
          } else {
            toast.error(
              `Failed to import event listener: ${res.error.message} (${res.error.code})`
            );
          }
        },
        onError: () => {
          toast.error(
            "Failed to import event listener, please check the share code and try again"
          );
        },
      }
    );
  }, [shareCode, eventListenersImportMutation]);

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
          <DialogTitle className="mb-1">Import Event Listener</DialogTitle>
          <DialogDescription>
            Import an event listener using a share code.
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
            disabled={eventListenersImportMutation.isPending}
          >
            {eventListenersImportMutation.isPending ? "Importing..." : "Import"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}