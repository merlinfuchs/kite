import { toast } from "sonner";
import { Card } from "../ui/card";
import { Button } from "../ui/button";
import { Skeleton } from "../ui/skeleton";
import AppEmptyPlaceholder from "../app/AppEmptyPlaceholder";
import {
  useMarketplaceReportResolveMutation,
  useMarketplaceReportsQuery,
} from "@/lib/api/marketplace";
import MarketplaceAuthor from "./MarketplaceAuthor";
import MarketplaceStatusBadge from "./MarketplaceStatusBadge";

export default function MarketplaceReportList({
  onSelect,
}: {
  onSelect: (listingId: string) => void;
}) {
  const query = useMarketplaceReportsQuery();
  const resolveMutation = useMarketplaceReportResolveMutation();

  if (query.isLoading) {
    return <Skeleton className="h-40" />;
  }

  const reports = query.data?.success
    ? query.data.data.flatMap((r) => (r ? [r] : []))
    : [];

  if (reports.length === 0) {
    return (
      <AppEmptyPlaceholder
        title="No open reports"
        description="Reports from users show up here."
      />
    );
  }

  function onDismiss(reportId: string) {
    resolveMutation.mutate(reportId, {
      onSuccess: (res) => {
        if (res.success) {
          toast.success("Report dismissed");
        } else {
          toast.error(
            `Failed to dismiss report: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  return (
    <div className="flex flex-col gap-3">
      {reports.map((report) => (
        <Card key={report.id} className="p-4">
          <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
            <div className="min-w-0 space-y-2">
              <div className="flex items-center gap-2">
                <div className="font-medium truncate">
                  {report.listing?.name ?? report.listing_id}
                </div>
                {report.listing && (
                  <MarketplaceStatusBadge status={report.listing.status} />
                )}
              </div>
              <p className="text-sm whitespace-pre-wrap">{report.reason}</p>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                Reported by <MarketplaceAuthor user={report.reporter} />
                on {new Date(report.created_at).toLocaleDateString()}
              </div>
            </div>
            <div className="flex gap-2 flex-none">
              <Button variant="ghost" onClick={() => onDismiss(report.id)}>
                Dismiss
              </Button>
              <Button
                variant="outline"
                onClick={() => onSelect(report.listing_id)}
              >
                Review listing
              </Button>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}
