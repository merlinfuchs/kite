import {
  BoxesIcon,
  DownloadIcon,
  MailPlusIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
  TriangleAlertIcon,
} from "lucide-react";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { Badge } from "../ui/badge";
import { MarketplaceListing } from "@/lib/types/wire.gen";
import {
  getRiskyBlocks,
  listingKindLabel,
  listingSummary,
} from "@/lib/marketplace";
import MarketplaceAuthor from "./MarketplaceAuthor";
import MarketplaceStatusBadge from "./MarketplaceStatusBadge";

const kindIcons = {
  command: SlashSquareIcon,
  event_listener: SatelliteDishIcon,
  message: MailPlusIcon,
  module: BoxesIcon,
};

export default function MarketplaceListingCard({
  listing,
  showStatus,
  onClick,
}: {
  listing: MarketplaceListing;
  showStatus?: boolean;
  onClick: () => void;
}) {
  const Icon =
    kindIcons[listing.kind as keyof typeof kindIcons] ?? kindIcons.module;
  const risky = getRiskyBlocks(listing.block_types).length > 0;

  return (
    <Card
      className="cursor-pointer hover:border-primary/60 transition-colors flex flex-col"
      onClick={onClick}
      role="button"
    >
      <CardHeader className="flex flex-row gap-4 p-4 space-y-0 flex-1">
        <div className="h-10 w-10 bg-primary/40 flex-none rounded-md flex items-center justify-center">
          <Icon className="w-6 h-6 text-primary" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 mb-1">
            <CardTitle className="text-base truncate">{listing.name}</CardTitle>
            {showStatus && <MarketplaceStatusBadge status={listing.status} />}
          </div>
          <CardDescription className="line-clamp-2 mb-3">
            {listing.description}
          </CardDescription>
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="outline">{listingKindLabel(listing.kind)}</Badge>
            {listing.kind === "module" && (
              <span className="text-xs text-muted-foreground">
                {listingSummary(listing)}
              </span>
            )}
            {risky && (
              <TriangleAlertIcon
                className="h-4 w-4 text-yellow-500"
                aria-label="Uses sensitive blocks"
              />
            )}
          </div>
        </div>
      </CardHeader>
      <div className="flex items-center justify-between px-4 pb-4 gap-3">
        <MarketplaceAuthor user={listing.author} />
        <div className="flex items-center gap-1 text-xs text-muted-foreground flex-none">
          <DownloadIcon className="h-3.5 w-3.5" />
          {listing.import_count}
        </div>
      </div>
    </Card>
  );
}
