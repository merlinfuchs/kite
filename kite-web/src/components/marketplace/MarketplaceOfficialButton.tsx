import { toast } from "sonner";
import { BadgeCheckIcon, BadgeXIcon } from "lucide-react";
import LoadingButton from "../common/LoadingButton";
import { useMarketplaceListingOfficialMutation } from "@/lib/api/marketplace";
import { MarketplaceListing } from "@/lib/types/wire.gen";

// Moderators can put published listings on the Official tab. They keep their
// author, only the Kite team's own templates are shown as made by Kite.
export default function MarketplaceOfficialButton({
  listing,
}: {
  listing: MarketplaceListing;
}) {
  const officialMutation = useMarketplaceListingOfficialMutation(listing.id);

  if (listing.status !== "approved" && !listing.official) {
    return null;
  }

  function toggle() {
    const official = !listing.official;
    officialMutation.mutate(
      { official },
      {
        onSuccess: (res) => {
          if (res.success) {
            toast.success(
              official
                ? "Listing is now official"
                : "Listing is no longer official"
            );
          } else {
            toast.error(
              `Failed to update listing: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <LoadingButton
      variant="outline"
      loading={officialMutation.isPending}
      onClick={toggle}
    >
      {listing.official ? (
        <BadgeXIcon className="h-4 w-4 mr-2" />
      ) : (
        <BadgeCheckIcon className="h-4 w-4 mr-2" />
      )}
      {listing.official ? "Remove official" : "Make official"}
    </LoadingButton>
  );
}
