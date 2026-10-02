import { useState } from "react";
import { toast } from "sonner";
import { CheckIcon, XIcon } from "lucide-react";
import { DialogFooter } from "../ui/dialog";
import { Textarea } from "../ui/textarea";
import { Label } from "../ui/label";
import LoadingButton from "../common/LoadingButton";
import { useMarketplaceListingReviewMutation } from "@/lib/api/marketplace";
import { MarketplaceListing } from "@/lib/types/wire.gen";

export default function MarketplaceReviewActions({
  listing,
  onDone,
}: {
  listing: MarketplaceListing;
  onDone: () => void;
}) {
  const [note, setNote] = useState("");
  const reviewMutation = useMarketplaceListingReviewMutation(listing.id);

  const takeDownStatus = listing.status === "approved" ? "removed" : "rejected";

  function review(status: "approved" | "rejected" | "removed") {
    if (status !== "approved" && !note.trim()) {
      toast.error("Add a note so the author knows what to fix");
      return;
    }

    reviewMutation.mutate(
      { status, note: note.trim() },
      {
        onSuccess: (res) => {
          if (res.success) {
            toast.success(
              status === "approved" ? "Listing approved" : "Listing taken down"
            );
            onDone();
          } else {
            toast.error(
              `Failed to review listing: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <div className="space-y-3 border-t pt-4">
      <div className="space-y-2">
        <Label htmlFor="review-note">Note for the author</Label>
        <Textarea
          id="review-note"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="Required when rejecting or removing, optional when approving."
          minRows={2}
          maxRows={6}
          maxLength={1000}
        />
      </div>
      <DialogFooter className="gap-2">
        <LoadingButton
          variant="destructive"
          loading={reviewMutation.isPending}
          onClick={() => review(takeDownStatus)}
        >
          <XIcon className="h-4 w-4 mr-2" />
          {takeDownStatus === "removed" ? "Remove" : "Reject"}
        </LoadingButton>
        {listing.status !== "approved" && (
          <LoadingButton
            loading={reviewMutation.isPending}
            onClick={() => review("approved")}
          >
            <CheckIcon className="h-4 w-4 mr-2" />
            Approve
          </LoadingButton>
        )}
      </DialogFooter>
    </div>
  );
}
