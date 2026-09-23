import { ReactNode, useState } from "react";
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
import { toast } from "sonner";

export default function FlowExportDialog({
  title,
  shareCode,
  children,
}: {
  title: string;
  shareCode: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  function copy() {
    navigator.clipboard
      .writeText(shareCode)
      .then(() => {
        toast.success("Copied to clipboard!");
        setOpen(false);
      })
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>
            Copy the share code below and import it into another app.
          </DialogDescription>
        </DialogHeader>
        <Textarea
          readOnly
          value={shareCode}
          className="min-h-[78px] max-h-[218px]"
          onFocus={(e) => e.target.select()}
        />
        <DialogFooter>
          <Button onClick={copy}>Copy to clipboard</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
