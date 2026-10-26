import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import BundleDialog from "@/components/app/BundleDialog";
import BundleExportDialog from "@/components/app/BundleExportDialog";
import BundleImportDialog from "@/components/app/BundleImportDialog";
import BundleListingDialog from "@/components/app/BundleListingDialog";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  useBundleDeleteMutation,
  useBundlePublishMutation,
} from "@/lib/api/mutations";
import env from "@/lib/env/client";
import { useApps, useBundleInstalls, useUserBundles } from "@/lib/hooks/api";
import { useTemplates } from "@/lib/hooks/templates";
import { useAppId } from "@/lib/hooks/params";
import { templatePath } from "@/lib/flow/templates";
import { UserBundle } from "@/lib/types/wire.gen";
import {
  CopyIcon,
  ImportIcon,
  PackageIcon,
  PlusIcon,
  ShareIcon,
  Trash2Icon,
} from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";
import { toast } from "sonner";

const breadcrumbs = [
  {
    label: "Bundles",
  },
];

export default function AppBundlesPage() {
  const appId = useAppId();
  const bundles = useUserBundles();

  // The ones made from this app first, they can get new versions here.
  const sorted = (bundles ?? [])
    .flatMap((b) => (b ? [b] : []))
    .sort(
      (a, b) =>
        Number(b.source_app_id === appId) - Number(a.source_app_id === appId)
    );

  return (
    <AppLayout title="Bundles" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row md:items-start justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">Bundles</h1>
          <p className="text-muted-foreground text-sm">
            Share commands, event listeners and message templates with other
            apps. Apps that add your bundle can update when you publish a new
            version.{" "}
            <a
              href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/templates#bundles`}
              target="_blank"
              className="text-primary hover:underline"
            >
              Learn More
            </a>
          </p>
        </div>
        <div className="flex gap-2 flex-none">
          <BundleImportDialog>
            <Button variant="outline">
              <ImportIcon className="mr-2 h-4 w-4" />
              Import
            </Button>
          </BundleImportDialog>
          <BundleExportDialog>
            <Button variant="outline">
              <ShareIcon className="mr-2 h-4 w-4" />
              Export
            </Button>
          </BundleExportDialog>
          <BundleDialog>
            <Button>
              <PlusIcon className="mr-2 h-4 w-4" />
              Create bundle
            </Button>
          </BundleDialog>
        </div>
      </div>
      <Separator className="my-8" />
      {bundles && !sorted.length && (
        <p className="text-sm text-muted-foreground">
          You haven&apos;t made a bundle yet. Create one from items of this app.
        </p>
      )}
      <div className="space-y-3">
        {sorted.map((bundle) => (
          <BundleEntry key={bundle.id} bundle={bundle} />
        ))}
      </div>
      <AddedBundles />
    </AppLayout>
  );
}

function BundleEntry({ bundle }: { bundle: UserBundle }) {
  const appId = useAppId();
  const apps = useApps();
  const deleteMutation = useBundleDeleteMutation();
  const publishMutation = useBundlePublishMutation();

  const items = bundle.item_count;
  const fromHere = bundle.source_app_id === appId;
  const sourceApp = apps?.find((a) => a?.id === bundle.source_app_id);
  const source = fromHere
    ? "From this app"
    : !bundle.source_app_id
    ? "Its app was deleted"
    : sourceApp
    ? `From ${sourceApp.name}`
    : "From an app you can't open anymore";

  function copyLink() {
    navigator.clipboard
      .writeText(
        `${env.NEXT_PUBLIC_SITE_URL}${
          bundle.listed_version !== null
            ? templatePath(bundle.id)
            : `/bundles/${bundle.id}`
        }`
      )
      .then(() => toast.success("Copied the link"))
      .catch(() => toast.error("Failed to copy, use ctrl + c instead"));
  }

  async function setPublished(published: boolean) {
    const res = await publishMutation.mutateAsync({
      bundleId: bundle.id,
      published,
    });
    if (!res.success) {
      toast.error(
        `Failed to ${published ? "share" : "stop sharing"} ${bundle.name}: ${
          res.error.message
        } (${res.error.code})`
      );
      return;
    }
    if (published) copyLink();
    else toast.success(`Stopped sharing ${bundle.name}`);
  }

  async function remove() {
    const res = await deleteMutation.mutateAsync(bundle.id);
    if (!res.success) {
      toast.error(
        `Failed to delete ${bundle.name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(`Deleted ${bundle.name}`);
  }

  return (
    <Card className="p-4 flex flex-col md:flex-row md:items-center gap-4">
      <Link
        href={{
          pathname: "/apps/[appId]/bundles/[bundleId]",
          query: { appId, bundleId: bundle.id },
        }}
        className="flex items-start gap-3 min-w-0 flex-1 group"
      >
        <div className="h-10 w-10 bg-primary/15 flex-none rounded-lg flex items-center justify-center">
          <PackageIcon className="h-5 w-5 text-primary" />
        </div>
        <div className="min-w-0">
          <div className="font-medium group-hover:underline truncate">
            {bundle.name}
          </div>
          {bundle.description && (
            <div className="text-sm text-muted-foreground line-clamp-2">
              {bundle.description}
            </div>
          )}
          <div className="text-xs text-muted-foreground mt-1">
            Version {bundle.version} · {items} {items === 1 ? "item" : "items"}{" "}
            · {source}
            {bundle.published && " · Shared"}
          </div>
          {bundle.removed_reason !== null ? (
            <RemovedStatus reason={bundle.removed_reason} />
          ) : (
            bundle.published && <ListingStatus bundle={bundle} />
          )}
        </div>
      </Link>
      <div className="flex gap-2 flex-none">
        {fromHere && (
          <BundleDialog bundle={bundle}>
            <Button variant="outline">Publish update</Button>
          </BundleDialog>
        )}
        {bundle.published && (
          <BundleListingDialog bundle={bundle}>
            <Button variant="outline">
              {bundle.listed_version !== null
                ? "Edit listing"
                : bundle.latest_review?.status === "pending"
                ? "Edit submission"
                : "List in marketplace"}
            </Button>
          </BundleListingDialog>
        )}
        {bundle.removed_reason !== null ? null : bundle.published ? (
          <>
            <Button variant="outline" onClick={copyLink}>
              <CopyIcon className="mr-2 h-4 w-4" />
              Copy link
            </Button>
            <ConfirmDialog
              title={`Stop sharing ${bundle.name}?`}
              description={
                bundle.listed_version !== null ||
                bundle.latest_review?.status === "pending"
                  ? "Its link stops working and it leaves the marketplace. Apps that added it keep what it added, but don't get updates anymore. Sharing it again needs a new review."
                  : "Its link stops working. Apps that added it keep what it added, but don't get updates anymore."
              }
              onConfirm={() => setPublished(false)}
            >
              <Button variant="outline" disabled={publishMutation.isPending}>
                Stop sharing
              </Button>
            </ConfirmDialog>
          </>
        ) : (
          <ConfirmDialog
            title={`Share ${bundle.name}?`}
            description="Anyone with its link can see what's inside and add it to their app. They get the versions you publish as updates. Secrets and the values of parameters aren't part of it. It isn't listed anywhere."
            onConfirm={() => setPublished(true)}
          >
            <Button variant="outline" disabled={publishMutation.isPending}>
              Share
            </Button>
          </ConfirmDialog>
        )}
        <ConfirmDialog
          title={`Delete ${bundle.name}?`}
          description="Apps that added it keep what it added, but don't get updates anymore. You can't undo this."
          onConfirm={remove}
        >
          <Button
            variant="outline"
            size="icon"
            disabled={deleteMutation.isPending}
          >
            <Trash2Icon className="h-4 w-4" />
          </Button>
        </ConfirmDialog>
      </div>
    </Card>
  );
}

// RemovedStatus says why Kite took the bundle down, which can only be appealed
// on Discord. It's part of the link to the bundle, so it can't link there.
function RemovedStatus({ reason }: { reason: string }) {
  return (
    <div className="text-xs text-destructive mt-1">
      Kite took this bundle down for breaking the rules: {reason} Only you can
      see it now. To appeal, ask on our Discord server.
    </div>
  );
}

// ListingStatus says where a shared bundle is in the marketplace review.
function ListingStatus({ bundle }: { bundle: UserBundle }) {
  const review = bundle.latest_review;
  const listed = bundle.listed_version;

  // A review of the listed version is a change of its listing.
  let status: ReactNode = null;
  if (review?.status === "pending") {
    status =
      listed === null
        ? "In review for the marketplace"
        : review.version === listed
        ? "Listed in the marketplace · Your changes to its listing are in review"
        : `Listed in the marketplace · Version ${review.version} is in review, apps get version ${listed} until then`;
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
    status = "Listed in the marketplace";
  }

  return (
    status && <div className="text-xs text-muted-foreground mt-1">{status}</div>
  );
}

// AddedBundles are the bundles the app added that aren't official or the
// user's, like ones other users made.
function AddedBundles() {
  const appId = useAppId();
  const installs = useBundleInstalls();
  const bundles = useUserBundles();
  const templates = useTemplates();
  if (!installs || !bundles || !templates) return null;

  const others = installs.flatMap((i) =>
    i &&
    !bundles.some((b) => b?.id === i.bundle_id) &&
    !templates.some((t) => t.id === i.bundle_id)
      ? [i]
      : []
  );
  if (!others.length) return null;

  return (
    <div className="mt-10">
      <h2 className="font-semibold mb-1">Added from others</h2>
      <p className="text-sm text-muted-foreground mb-4">
        Bundles you didn&apos;t make, or deleted, that this app added.
      </p>
      <div className="space-y-3">
        {others.map((install) => (
          <Card key={install.id} className="p-4">
            <Link
              href={{
                pathname: "/apps/[appId]/bundles/[bundleId]",
                query: { appId, bundleId: install.bundle_id },
              }}
              className="flex items-start gap-3 group"
            >
              <div className="h-10 w-10 bg-muted flex-none rounded-lg flex items-center justify-center">
                <PackageIcon className="h-5 w-5 text-muted-foreground" />
              </div>
              <div className="min-w-0">
                <div className="font-medium group-hover:underline truncate">
                  {install.name}
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  Version {install.bundle_version}
                </div>
              </div>
            </Link>
          </Card>
        ))}
      </div>
    </div>
  );
}

AppBundlesPage.getLayout = getAppShellLayout;
