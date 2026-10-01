import { ReactNode, useEffect, useState } from "react";
import { toast } from "sonner";
import {
  DownloadIcon,
  EyeIcon,
  EyeOffIcon,
  FlagIcon,
  PencilIcon,
  MailPlusIcon,
  SatelliteDishIcon,
  SlashSquareIcon,
  TrashIcon,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Badge } from "../ui/badge";
import { Card } from "../ui/card";
import { Switch } from "../ui/switch";
import { Skeleton } from "../ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import LoadingButton from "../common/LoadingButton";
import ConfirmDialog from "../common/ConfirmDialog";
import {
  useMarketplaceListingDeleteMutation,
  useMarketplaceListingQuery,
  useMarketplaceMeQuery,
} from "@/lib/api/marketplace";
import { useUser } from "@/lib/hooks/api";
import { MarketplaceListing } from "@/lib/types/wire.gen";
import { listingKindLabel, listingStatusLabel } from "@/lib/marketplace";
import MarketplaceAuthor from "./MarketplaceAuthor";
import MarketplaceStatusBadge from "./MarketplaceStatusBadge";
import MarketplaceRiskAlert from "./MarketplaceRiskAlert";
import MarketplaceFlowViewer from "./MarketplaceFlowViewer";
import MarketplaceMessageViewer from "./MarketplaceMessageViewer";
import MarketplaceReportDialog from "./MarketplaceReportDialog";
import MarketplacePublishDialog from "./MarketplacePublishDialog";
import { useMarketplaceImport } from "./useMarketplaceImport";

export default function MarketplaceListingDialog({
  listingId,
  onClose,
  actions,
}: {
  listingId: string | null;
  onClose: () => void;
  // Replaces the import footer, used by the moderation page for review controls.
  actions?: (listing: MarketplaceListing) => ReactNode;
}) {
  return (
    <Dialog open={!!listingId} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="overflow-y-auto max-h-[90dvh] max-w-3xl">
        {listingId && (
          <ListingDetails
            listingId={listingId}
            onClose={onClose}
            actions={actions}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

// Separate from the dialog so the listing and import state reset when it's reopened.
function ListingDetails({
  listingId,
  onClose,
  actions,
}: {
  listingId: string;
  onClose: () => void;
  actions?: (listing: MarketplaceListing) => ReactNode;
}) {
  const query = useMarketplaceListingQuery(listingId);
  const me = useMarketplaceMeQuery();
  const user = useUser();

  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [previewing, setPreviewing] = useState<number | null>(null);
  const [editing, setEditing] = useState(false);

  const { importListing, ready, pending } = useMarketplaceImport();
  const deleteMutation = useMarketplaceListingDeleteMutation();

  const listing = query.data?.success ? query.data.data : undefined;

  useEffect(() => {
    if (listing) {
      setSelected(new Set(listing.items.map((_, i) => i)));
    }
  }, [listing]);

  if (query.data && !query.data.success) {
    return (
      <DialogHeader>
        <DialogTitle>Listing not found</DialogTitle>
        <DialogDescription>
          This listing was deleted or isn&apos;t public anymore.
        </DialogDescription>
      </DialogHeader>
    );
  }

  if (!listing) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-7 w-1/2" />
        <Skeleton className="h-20 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  const isAuthor = !!user && listing.author?.id === user.id;
  const isModerator = !!(me.data?.success && me.data.data.is_moderator);

  function onImport() {
    if (!listing) return;
    if (selected.size === 0) {
      toast.error("Select at least one item to import");
      return;
    }
    importListing(listing.id, selected).then((ok) => {
      if (ok) onClose();
    });
  }

  function onDelete() {
    if (!listing) return;
    deleteMutation.mutate(listing.id, {
      onSuccess: (res) => {
        if (res.success) {
          toast.success("Listing deleted");
          onClose();
        } else {
          toast.error(
            `Failed to delete listing: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }

  function toggle(i: number, checked: boolean) {
    const next = new Set(selected);
    if (checked) next.add(i);
    else next.delete(i);
    setSelected(next);
  }

  return (
    <>
      <DialogHeader>
        <div className="flex flex-wrap items-center gap-2 mb-1">
          <DialogTitle className="text-xl">{listing.name}</DialogTitle>
          <Badge variant="outline">{listingKindLabel(listing.kind)}</Badge>
          {listing.status !== "approved" && (
            <MarketplaceStatusBadge status={listing.status} />
          )}
        </div>
        <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
          <MarketplaceAuthor user={listing.author} />
          <span className="text-sm text-muted-foreground">
            {listing.import_count} import
            {listing.import_count === 1 ? "" : "s"}
          </span>
          <span className="text-sm text-muted-foreground">
            Updated {new Date(listing.updated_at).toLocaleDateString()}
          </span>
        </div>
      </DialogHeader>

      <div className="space-y-5 my-2">
        {(isAuthor || isModerator) && listing.review_note && (
          <Alert
            variant={listing.status === "approved" ? "default" : "destructive"}
          >
            <AlertTitle>
              {listingStatusLabel(listing.status)}: note from the moderators
            </AlertTitle>
            <AlertDescription>{listing.review_note}</AlertDescription>
          </Alert>
        )}
        {isAuthor && listing.status === "pending" && !listing.review_note && (
          <Alert>
            <AlertTitle>Waiting for review</AlertTitle>
            <AlertDescription className="text-muted-foreground">
              Only you and the moderators can see this listing until it&apos;s
              approved.
            </AlertDescription>
          </Alert>
        )}

        <p className="text-sm whitespace-pre-wrap">{listing.description}</p>

        <MarketplaceRiskAlert blockTypes={listing.block_types} />

        <div>
          <div className="font-medium mb-0.5">Contents</div>
          <div className="text-sm text-muted-foreground mb-3">
            {actions
              ? "Open the preview of every item and check what its blocks do."
              : "Select the items you want to import into this app."}
          </div>
          <div className="flex flex-col gap-3">
            {listing.items.map((item, i) => (
              <Card className="px-4 pb-4 pt-3" key={i}>
                <div className="flex items-start gap-3">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-1.5 mb-0.5">
                      {item.type === "command" ? (
                        <SlashSquareIcon className="text-muted-foreground h-5 w-5 flex-none" />
                      ) : item.type === "message" ? (
                        <MailPlusIcon className="text-muted-foreground h-5 w-5 flex-none" />
                      ) : (
                        <SatelliteDishIcon className="text-muted-foreground h-5 w-5 flex-none" />
                      )}
                      <div className="font-medium truncate">
                        {item.type === "command" ? `/${item.name}` : item.name}
                      </div>
                      {item.source === "schedule" && (
                        <Badge variant="secondary">Scheduled</Badge>
                      )}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {item.description}
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setPreviewing(previewing === i ? null : i)}
                  >
                    {previewing === i ? (
                      <EyeOffIcon className="h-4 w-4 mr-1.5" />
                    ) : (
                      <EyeIcon className="h-4 w-4 mr-1.5" />
                    )}
                    Preview
                  </Button>
                  {!actions && (
                    <Switch
                      className="mt-1"
                      checked={selected.has(i)}
                      onCheckedChange={(checked) => toggle(i, checked)}
                    />
                  )}
                </div>
                {previewing === i && item.type === "message" && (
                  <div className="mt-4">
                    <MarketplaceMessageViewer item={item} />
                  </div>
                )}
                {previewing === i &&
                  item.type !== "message" &&
                  item.flow_source && (
                    <div className="mt-4">
                      <MarketplaceFlowViewer
                        flow={item.flow_source}
                        context={
                          item.type === "command"
                            ? "command"
                            : item.source === "schedule"
                            ? "event_schedule"
                            : "event_discord"
                        }
                      />
                    </div>
                  )}
              </Card>
            ))}
          </div>
        </div>
      </div>

      {actions ? (
        actions(listing)
      ) : (
        <DialogFooter className="gap-2 sm:justify-between sm:space-x-0">
          <div className="flex flex-wrap gap-2">
            {!isAuthor && listing.status === "approved" && (
              <MarketplaceReportDialog listingId={listing.id}>
                <Button variant="ghost">
                  <FlagIcon className="h-4 w-4 mr-2" />
                  Report
                </Button>
              </MarketplaceReportDialog>
            )}
            {isAuthor && (
              <Button variant="outline" onClick={() => setEditing(true)}>
                <PencilIcon className="h-4 w-4 mr-2" />
                Edit
              </Button>
            )}
            {(isAuthor || isModerator) && (
              <ConfirmDialog
                title="Delete listing?"
                description="The listing is removed from the marketplace. Commands and event listeners that were already imported stay in their apps."
                onConfirm={onDelete}
              >
                <Button variant="outline">
                  <TrashIcon className="h-4 w-4 mr-2" />
                  Delete
                </Button>
              </ConfirmDialog>
            )}
          </div>
          <LoadingButton onClick={onImport} loading={pending || !ready}>
            <DownloadIcon className="h-4 w-4 mr-2" />
            Import{" "}
            {selected.size === listing.items.length ? "all" : selected.size}
          </LoadingButton>
        </DialogFooter>
      )}

      {isAuthor && (
        <MarketplacePublishDialog
          listing={listing}
          open={editing}
          onOpenChange={setEditing}
        />
      )}
    </>
  );
}
