import { useState } from "react";
import AppLayout from "@/components/app/AppLayout";
import AppEmptyPlaceholder from "@/components/app/AppEmptyPlaceholder";
import MarketplaceListingDialog from "@/components/marketplace/MarketplaceListingDialog";
import MarketplaceListingGrid from "@/components/marketplace/MarketplaceListingGrid";
import MarketplaceModeratorList from "@/components/marketplace/MarketplaceModeratorList";
import MarketplaceReportList from "@/components/marketplace/MarketplaceReportList";
import MarketplaceReviewActions from "@/components/marketplace/MarketplaceReviewActions";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  MarketplaceListingStatus,
  useMarketplaceMeQuery,
  useMarketplaceModerationListingsQuery,
} from "@/lib/api/marketplace";

const breadcrumbs = [
  {
    label: "Marketplace",
    href: "/apps/[appId]/marketplace",
  },
  {
    label: "Moderation",
  },
];

export default function AppMarketplaceModerationPage() {
  const me = useMarketplaceMeQuery();
  const [listingId, setListingId] = useState<string | null>(null);

  const isModerator = !!(me.data?.success && me.data.data.is_moderator);
  const isAdmin = !!(me.data?.success && me.data.data.is_admin);

  return (
    <AppLayout title="Marketplace Moderation" breadcrumbs={breadcrumbs}>
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Moderation</h1>
        <p className="text-muted-foreground text-sm">
          Review listings before they go public and handle reports. Open the
          preview of every item and check what its blocks do before approving.
        </p>
      </div>
      <Separator className="my-8" />

      {me.isLoading ? (
        <Skeleton className="h-40" />
      ) : !isModerator ? (
        <AppEmptyPlaceholder
          title="Moderators only"
          description="You need to be a marketplace moderator to see this page."
        />
      ) : (
        <Tabs defaultValue="queue">
          <TabsList className="mb-6">
            <TabsTrigger value="queue">Queue</TabsTrigger>
            <TabsTrigger value="reports">Reports</TabsTrigger>
            <TabsTrigger value="history">All Listings</TabsTrigger>
            <TabsTrigger value="moderators">Moderators</TabsTrigger>
          </TabsList>
          <TabsContent value="queue">
            <StatusListings status="pending" onSelect={setListingId} />
          </TabsContent>
          <TabsContent value="reports">
            <MarketplaceReportList onSelect={setListingId} />
          </TabsContent>
          <TabsContent value="history">
            <HistoryTab onSelect={setListingId} />
          </TabsContent>
          <TabsContent value="moderators">
            <MarketplaceModeratorList isAdmin={isAdmin} />
          </TabsContent>
        </Tabs>
      )}

      <MarketplaceListingDialog
        listingId={listingId}
        onClose={() => setListingId(null)}
        actions={(listing) => (
          <MarketplaceReviewActions
            listing={listing}
            onDone={() => setListingId(null)}
          />
        )}
      />
    </AppLayout>
  );
}

function StatusListings({
  status,
  onSelect,
}: {
  status: MarketplaceListingStatus;
  onSelect: (id: string) => void;
}) {
  const query = useMarketplaceModerationListingsQuery(status);

  return (
    <MarketplaceListingGrid
      listings={query.data?.success ? query.data.data : undefined}
      loading={query.isLoading}
      showStatus
      emptyTitle={status === "pending" ? "Queue is empty" : "No listings"}
      emptyDescription={
        status === "pending"
          ? "New and changed listings show up here."
          : "There are no listings with this status."
      }
      onSelect={onSelect}
    />
  );
}

function HistoryTab({ onSelect }: { onSelect: (id: string) => void }) {
  const [status, setStatus] = useState<MarketplaceListingStatus>("approved");

  return (
    <div className="space-y-6">
      <Select
        value={status}
        onValueChange={(v) => setStatus(v as MarketplaceListingStatus)}
      >
        <SelectTrigger className="sm:w-48">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="approved">Published</SelectItem>
          <SelectItem value="rejected">Rejected</SelectItem>
          <SelectItem value="removed">Removed</SelectItem>
        </SelectContent>
      </Select>
      <StatusListings status={status} onSelect={onSelect} />
    </div>
  );
}
