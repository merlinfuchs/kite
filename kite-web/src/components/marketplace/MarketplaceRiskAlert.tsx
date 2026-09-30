import { GlobeIcon, TriangleAlertIcon } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { getRiskyBlocks } from "@/lib/marketplace";

// Lists the blocks that can do damage or send data out of Discord, so users
// know what they're importing.
export default function MarketplaceRiskAlert({
  blockTypes,
}: {
  blockTypes: string[];
}) {
  const risky = getRiskyBlocks(blockTypes);
  if (!risky.length) {
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
    </div>
  );
}
