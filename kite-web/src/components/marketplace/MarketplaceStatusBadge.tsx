import { Badge } from "../ui/badge";
import { listingStatusLabel } from "@/lib/marketplace";

export default function MarketplaceStatusBadge({ status }: { status: string }) {
  return (
    <Badge
      variant={
        status === "approved"
          ? "default"
          : status === "pending"
          ? "secondary"
          : "destructive"
      }
    >
      {listingStatusLabel(status)}
    </Badge>
  );
}
