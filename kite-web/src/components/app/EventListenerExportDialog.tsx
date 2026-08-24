import { ReactNode, useMemo, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import { EventListener } from "@/lib/types/wire.gen";
import { toast } from "sonner";

export function EventListenerExportDialog({
  listener,
  children,
}: {
  listener: EventListener;
  children: ReactNode;
}) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const shareCode = useMemo(() => {
    try {
      return JSON.stringify({
        source: listener.source,
        type: listener.type,
        description: listener.description,
        flow_source: listener.flow_source,
        enabled: listener.enabled,
      });
    } catch {
      return null;
    }
  }, [listener]);

  const handleCopy = () => {
    if (!shareCode) {
      toast.error("Failed to export event listener: error generating share code");
      return;
    }

    if (!navigator.clipboard || !navigator.clipboard.writeText) {
      toast.error("Failed to copy event listener: use ctrl + c to copy");
      return;
    }

    navigator.clipboard
      .writeText(shareCode)
      .then(() => {
        toast.success("Copied to clipboard!");
        setDialogOpen(false);
      })
      .catch(() => {
        toast.error("Failed to copy event listener: use ctrl + c to copy");
      });
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="mb-1">Export Event Listener</DialogTitle>
          <DialogDescription>
            Share your event listener using the code below.
          </DialogDescription>
        </DialogHeader>
        <div className="mb-3 mt-2 space-y-3">
          <div>
            <div className="font-medium mb-0.5">Share Code</div>
            <Textarea
              readOnly
              value={shareCode ?? ""}
              className="min-h-[78px] max-h-[218px] overflow-y-auto resize-y"
            />
          </div>
        </div>
        <DialogFooter>
          <Button onClick={handleCopy} disabled={!shareCode}>
            Copy To Clipboard
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}