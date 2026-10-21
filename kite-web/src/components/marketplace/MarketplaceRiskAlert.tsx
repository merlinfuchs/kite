import { GlobeIcon, KeyIcon, TriangleAlertIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import {
  getListingReferencedSecretNames,
  getRiskyBlocks,
} from "@/lib/marketplace";
import { MarketplaceListingItem } from "@/lib/types/wire.gen";
import { useMemo } from "react";

// Lists the blocks that can do damage, send data out of Discord, or require
// secrets, so users know what they're importing.
export default function MarketplaceRiskAlert({
  blockTypes,
  items,
}: {
  blockTypes: string[];
  items?: MarketplaceListingItem[];
}) {
  const risky = getRiskyBlocks(blockTypes);
  const referencedSecrets = useMemo(
    () => getListingReferencedSecretNames(items),
    [items]
  );

  if (!risky.length && !referencedSecrets.length) {
    return null;
  }

  const destructive = risky.filter((b) => b.risk === "destructive");
  const external = risky.filter((b) => b.risk === "external");

  return (
    <div className="space-y-3">
      {!!destructive.length && (
        <Alert>
          <TriangleAlertIcon className="h-4 w-4" />
          <AlertTitle>Can change or remove things in your server</AlertTitle>
          <AlertDescription className="text-muted-foreground">
            Uses {destructive.map((b) => b.title).join(", ")}. Check who can run
            these before enabling them.
          </AlertDescription>
        </Alert>
      )}
      {!!external.length && (
        <Alert>
          <GlobeIcon className="h-4 w-4" />
          <AlertTitle>Can send data outside of Discord</AlertTitle>
          <AlertDescription className="text-muted-foreground">
            Uses {external.map((b) => b.title).join(", ")}. Open the preview and
            check where the requests go before importing.
          </AlertDescription>
        </Alert>
      )}
      {!!referencedSecrets.length && (
        <Alert>
          <KeyIcon className="h-4 w-4" />
          <AlertTitle>Requires App Secrets</AlertTitle>
          <AlertDescription className="text-muted-foreground">
            This listing uses secret variables (
            {referencedSecrets.map((s) => `{{secrets.${s}}}`).join(", ")}).
            Verify where requests send data before supplying secrets or
            enabling these flows.
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
