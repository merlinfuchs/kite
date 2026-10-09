import { ReactNode, useState } from "react";
import { toast } from "sonner";
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
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { useMarketplaceReportCreateMutation } from "@/lib/api/marketplace";

export default function MarketplaceReportDialog({
  listingId,
  children,
}: {
  listingId: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState("");

  const reportMutation = useMarketplaceReportCreateMutation(listingId);

  function onReport() {
    if (reason.trim().length < 5) {
      toast.error("Tell the moderators what's wrong with this listing");
      return;
    }

    reportMutation.mutate(
      { reason: reason.trim() },
      {
        onSuccess: (res) => {
          if (res.success) {
            toast.success("Thanks, a moderator will take a look");
            setReason("");
            setOpen(false);
          } else {
            toast.error(
              `Failed to report listing: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Report listing</DialogTitle>
          <DialogDescription>
            Report listings that are malicious, broken, stolen or break the
            rules. Listings with several reports are hidden until a moderator
            reviews them.
          </DialogDescription>
        </DialogHeader>
        <Textarea
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="What's wrong with this listing?"
          minRows={4}
          maxRows={8}
          maxLength={1000}
          autoFocus
        />
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <LoadingButton
            variant="destructive"
            onClick={onReport}
            loading={reportMutation.isPending}
          >
            Report
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
