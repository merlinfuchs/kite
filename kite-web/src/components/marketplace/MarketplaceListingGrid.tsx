import { MarketplaceListing } from "@/lib/types/wire.gen";
import { Skeleton } from "../ui/skeleton";
import AppEmptyPlaceholder from "../app/AppEmptyPlaceholder";
import MarketplaceListingCard from "./MarketplaceListingCard";

export default function MarketplaceListingGrid({
  listings,
  loading,
  showStatus,
  emptyTitle,
  emptyDescription,
  onSelect,
}: {
  listings?: (MarketplaceListing | undefined)[];
  loading: boolean;
  showStatus?: boolean;
  emptyTitle: string;
  emptyDescription: string;
  onSelect: (listingId: string) => void;
}) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-40" />
        ))}
      </div>
    );
  }

  const items = listings?.flatMap((l) => (l ? [l] : [])) ?? [];
  if (items.length === 0) {
    return (
      <AppEmptyPlaceholder title={emptyTitle} description={emptyDescription} />
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
      {items.map((listing) => (
        <MarketplaceListingCard
          key={listing.id}
          listing={listing}
          showStatus={showStatus}
          onClick={() => onSelect(listing.id)}
        />
      ))}
    </div>
  );
}
