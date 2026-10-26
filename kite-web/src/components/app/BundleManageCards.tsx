import ConfirmDialog from "@/components/common/ConfirmDialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import {
  useBundleDeleteMutation,
  useBundleListingDeleteMutation,
} from "@/lib/api/mutations";
import env from "@/lib/env/client";
import { templatePath } from "@/lib/flow/templates";
import { useAppId } from "@/lib/hooks/params";
import { UserBundle } from "@/lib/types/wire.gen";
import { CopyIcon } from "lucide-react";
import { useRouter } from "next/router";
import { ReactNode } from "react";
import { toast } from "sonner";
import BundleDialog from "./BundleDialog";
import BundleListingDialog from "./BundleListingDialog";

// bundleUrl is where others see the bundle: its page in the marketplace while
// it's listed, otherwise the page of its link.
export function bundleUrl(bundle: UserBundle) {
  return `${env.NEXT_PUBLIC_SITE_URL}${
    bundle.listed_version !== null
      ? templatePath(bundle.id)
      : `/bundles/${bundle.id}`
  }`;
}

// BundleManageCards are where the user shares, lists and deletes a bundle
// they made, on its page. A bundle Kite took down can't be deleted, so apps
// that added it keep the warning.
export default function BundleManageCards({ bundle }: { bundle: UserBundle }) {
  return (
    <>
      {bundle.removed_reason !== null ? (
        <Card className="p-5">
          <div className="font-semibold mb-1">Taken down</div>
          <RemovedStatus reason={bundle.removed_reason} />
        </Card>
      ) : (
        <>
          <MarketplaceCard bundle={bundle} />
          <LinkCard bundle={bundle} />
          <DeleteCard bundle={bundle} />
        </>
      )}
    </>
  );
}

// SourceAppCard is for the app a bundle is made from, which has the items it
// was made of rather than a copy.
export function SourceAppCard({ bundle }: { bundle: UserBundle }) {
  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">Made from this app</div>
      <p className="text-sm text-muted-foreground mb-4">
        This app has the commands, event listeners and messages the template is
        made of. Change them here, then publish an update so apps that added the
        template get your changes.
      </p>
      <BundleDialog bundle={bundle}>
        <Button className="w-full">Publish update</Button>
      </BundleDialog>
    </Card>
  );
}

function LinkCard({ bundle }: { bundle: UserBundle }) {
  function copyLink() {
    navigator.clipboard
      .writeText(bundleUrl(bundle))
      .then(() => toast.success("Copied the link"))
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">Link</div>
      <p className="text-sm text-muted-foreground mb-4">
        Anyone with the link can see what&apos;s inside and add it to their app.
        Secrets and the values of parameters aren&apos;t part of it.
      </p>
      <Button variant="outline" className="w-full" onClick={copyLink}>
        <CopyIcon className="mr-2 h-4 w-4" />
        Copy link
      </Button>
    </Card>
  );
}

function MarketplaceCard({ bundle }: { bundle: UserBundle }) {
  const unlistMutation = useBundleListingDeleteMutation();
  const inMarketplace =
    bundle.listed_version !== null ||
    bundle.latest_review?.status === "pending";

  async function unlist() {
    const res = await unlistMutation.mutateAsync(bundle.id);
    if (!res.success) {
      toast.error(
        `Failed to remove ${bundle.name} from the marketplace: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(`Removed ${bundle.name} from the marketplace`);
  }

  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">Marketplace</div>
      <div className="text-sm text-muted-foreground mb-4">
        <ListingStatus bundle={bundle} /> Kite reviews templates before
        they&apos;re listed, and each new version of a listed one.
      </div>
      <div className="flex flex-col gap-2">
        <BundleListingDialog bundle={bundle}>
          <Button
            variant={inMarketplace ? "outline" : "default"}
            className="w-full"
          >
            {bundle.listed_version !== null
              ? "Edit listing"
              : bundle.latest_review?.status === "pending"
              ? "Edit submission"
              : "List in marketplace"}
          </Button>
        </BundleListingDialog>
        {inMarketplace && (
          <ConfirmDialog
            title={`Remove ${bundle.name} from the marketplace?`}
            description="Apps that added it stay on the last version Kite approved. Listing it again needs a new review."
            onConfirm={unlist}
          >
            <Button
              variant="ghost"
              className="w-full"
              disabled={unlistMutation.isPending}
            >
              Remove from marketplace
            </Button>
          </ConfirmDialog>
        )}
      </div>
    </Card>
  );
}

function DeleteCard({ bundle }: { bundle: UserBundle }) {
  const appId = useAppId();
  const router = useRouter();
  const deleteMutation = useBundleDeleteMutation();

  async function remove() {
    const res = await deleteMutation.mutateAsync(bundle.id);
    if (!res.success) {
      toast.error(
        `Failed to delete ${bundle.name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(`Deleted ${bundle.name}`);
    router.push({ pathname: "/apps/[appId]/bundles", query: { appId } });
  }

  return (
    <ConfirmDialog
      title={`Delete ${bundle.name}?`}
      description="Its link stops working and it leaves the marketplace. Apps that added it keep what it added, but don't get updates anymore. You can't undo this."
      onConfirm={remove}
    >
      <Button
        variant="ghost"
        className="w-full text-destructive hover:text-destructive"
        disabled={deleteMutation.isPending}
      >
        Delete template
      </Button>
    </ConfirmDialog>
  );
}

// ListingBadge shows at a glance whether a bundle is in the marketplace.
export function ListingBadge({ bundle }: { bundle: UserBundle }) {
  const review = bundle.latest_review;
  const [label, className] =
    bundle.removed_reason !== null
      ? ["Taken down", "border-transparent bg-destructive/15 text-destructive"]
      : bundle.listed_version !== null
      ? [
          "Listed",
          "border-transparent bg-green-500/15 text-green-600 dark:text-green-500",
        ]
      : review?.status === "pending"
      ? [
          "In review",
          "border-transparent bg-amber-500/15 text-amber-600 dark:text-amber-500",
        ]
      : review?.status === "rejected"
      ? [
          "Not approved",
          "border-transparent bg-destructive/15 text-destructive",
        ]
      : ["Not listed", "text-muted-foreground"];

  return (
    <Badge variant="outline" className={cn("flex-none font-medium", className)}>
      {label}
    </Badge>
  );
}

// RemovedStatus says why Kite took the bundle down, which can only be appealed
// on Discord.
export function RemovedStatus({ reason }: { reason: string }) {
  return (
    <span className="text-destructive">
      Kite took this template down for breaking the rules: {reason} Only you can
      see it now. To appeal, ask on our Discord server.
    </span>
  );
}

// ListingStatus says where a bundle is in the marketplace review.
export function ListingStatus({ bundle }: { bundle: UserBundle }) {
  const review = bundle.latest_review;
  const listed = bundle.listed_version;

  // A review of the listed version is a change of its listing.
  let status: ReactNode = "Not in the marketplace.";
  if (review?.status === "pending") {
    status =
      listed === null
        ? "In review for the marketplace."
        : review.version === listed
        ? "Listed in the marketplace, your changes to its listing are in review."
        : `Listed in the marketplace, version ${review.version} is in review and apps get version ${listed} until then.`;
  } else if (review?.status === "rejected") {
    status = (
      <span className="text-destructive">
        {listed === null
          ? "Not approved for the marketplace: "
          : review.version === listed
          ? "Listed in the marketplace, but your changes to its listing weren't approved: "
          : `Version ${review.version} wasn't approved, apps keep version ${listed}: `}
        {review.reason}
      </span>
    );
  } else if (listed !== null) {
    status = "Listed in the marketplace.";
  } else if (bundle.approved_version !== null) {
    status = `Not in the marketplace. Apps get version ${bundle.approved_version}, the last one Kite approved, until it's listed again.`;
  }
  return <>{status}</>;
}
