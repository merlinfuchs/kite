import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import BundleDialog from "@/components/app/BundleDialog";
import BundleExportDialog from "@/components/app/BundleExportDialog";
import BundleImportDialog from "@/components/app/BundleImportDialog";
import BundleListingDialog from "@/components/app/BundleListingDialog";
import {
  ListingBadge,
  ListingStatus,
  RemovedStatus,
} from "@/components/app/BundleManageCards";
import InstalledBundleList from "@/components/app/InstalledBundleList";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import env from "@/lib/env/client";
import { useApps, useBundleInstalls, useUserBundles } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { UserBundle } from "@/lib/types/wire.gen";
import {
  ChevronRightIcon,
  ImportIcon,
  PackageIcon,
  PlusIcon,
  ShareIcon,
} from "lucide-react";
import Link from "next/link";

const breadcrumbs = [
  {
    label: "Your Templates",
  },
];

export default function AppBundlesPage() {
  const appId = useAppId();
  const bundles = useUserBundles();
  const installs = useBundleInstalls()?.flatMap((i) => (i ? [i] : []));

  // The ones made from this app first, they can get new versions here.
  const sorted = (bundles ?? [])
    .flatMap((b) => (b ? [b] : []))
    .sort(
      (a, b) =>
        Number(b.source_app_id === appId) - Number(a.source_app_id === appId)
    );

  return (
    <AppLayout title="Your Templates" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row md:items-start justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            Your Templates
          </h1>
          <p className="text-muted-foreground text-sm">
            The templates this app added, and the ones you made from its
            commands, event listeners and message templates to use in other apps
            or share.{" "}
            <a
              href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/templates#your-own-templates`}
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
              Create template
            </Button>
          </BundleDialog>
        </div>
      </div>
      <Separator className="my-8" />
      {installs && bundles && !installs.length && !sorted.length && (
        <EmptyBundles />
      )}
      {!!installs?.length && (
        <section className="mb-10">
          <h2 className="font-semibold mb-1">Added to this app</h2>
          <p className="text-sm text-muted-foreground mb-4">
            Open one to update it, change its parameters or remove it.
          </p>
          <InstalledBundleList installs={installs} />
        </section>
      )}
      {!!sorted.length && (
        <section>
          <h2 className="font-semibold mb-1">Made by you</h2>
          <p className="text-sm text-muted-foreground mb-4">
            Apps that add your template can update when you publish a new
            version.
          </p>
          <div className="space-y-3">
            {sorted.map((bundle) => (
              <BundleEntry key={bundle.id} bundle={bundle} />
            ))}
          </div>
        </section>
      )}
    </AppLayout>
  );
}

// EmptyBundles explains bundles to apps that haven't added or made one.
function EmptyBundles() {
  const appId = useAppId();

  return (
    <Card className="p-8 flex flex-col items-center text-center">
      <div className="h-12 w-12 bg-primary/15 rounded-xl flex items-center justify-center mb-4">
        <PackageIcon className="h-6 w-6 text-primary" />
      </div>
      <div className="font-semibold mb-1">Nothing here yet</div>
      <p className="text-sm text-muted-foreground max-w-md mb-5">
        Templates you add from the Marketplace show up here, where you can
        update them. You can also make a template from commands, event listeners
        and messages of this app to add it to your other apps or share it.
      </p>
      <div className="flex flex-wrap justify-center gap-2">
        <Button asChild>
          <Link
            href={{ pathname: "/apps/[appId]/marketplace", query: { appId } }}
          >
            Browse the Marketplace
          </Link>
        </Button>
        <BundleDialog>
          <Button variant="outline">Create a template</Button>
        </BundleDialog>
      </div>
    </Card>
  );
}

// BundleEntry is a bundle the user made. Its page is where it's updated,
// shared, listed and deleted.
function BundleEntry({ bundle }: { bundle: UserBundle }) {
  const appId = useAppId();
  const apps = useApps();

  const items = bundle.item_count;
  const sourceApp = apps?.find((a) => a?.id === bundle.source_app_id);
  const source =
    bundle.source_app_id === appId
      ? "From this app"
      : !bundle.source_app_id
      ? "Its app was deleted"
      : sourceApp
      ? `From ${sourceApp.name}`
      : "From an app you can't open anymore";
  const review = bundle.latest_review;
  const inMarketplace =
    bundle.listed_version !== null || review?.status === "pending";
  // Details beyond the badge, like what's in review or why it wasn't approved.
  const details =
    bundle.removed_reason !== null ||
    review?.status === "rejected" ||
    (review?.status === "pending" && bundle.listed_version !== null);

  return (
    <Card className="p-4 flex items-center gap-3">
      <Link
        href={{
          pathname: "/apps/[appId]/bundles/[bundleId]",
          query: { appId, bundleId: bundle.id },
        }}
        className="flex items-center gap-3 group min-w-0 flex-1"
      >
        <div className="h-10 w-10 bg-primary/15 flex-none rounded-lg flex items-center justify-center self-start">
          <PackageIcon className="h-5 w-5 text-primary" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 min-w-0">
            <span className="font-medium group-hover:underline truncate">
              {bundle.name}
            </span>
            <ListingBadge bundle={bundle} />
          </div>
          {bundle.description && (
            <div className="text-sm text-muted-foreground line-clamp-2">
              {bundle.description}
            </div>
          )}
          <div className="text-xs text-muted-foreground mt-1">
            Version {bundle.version} · {items} {items === 1 ? "item" : "items"}{" "}
            · {source}
          </div>
          {details && (
            <div className="text-xs text-muted-foreground mt-1">
              {bundle.removed_reason !== null ? (
                <RemovedStatus reason={bundle.removed_reason} />
              ) : (
                <ListingStatus bundle={bundle} />
              )}
            </div>
          )}
        </div>
      </Link>
      {!inMarketplace && bundle.removed_reason === null && (
        <BundleListingDialog bundle={bundle}>
          <Button variant="outline" size="sm" className="flex-none">
            List in marketplace
          </Button>
        </BundleListingDialog>
      )}
      <ChevronRightIcon className="h-4 w-4 flex-none text-muted-foreground" />
    </Card>
  );
}

AppBundlesPage.getLayout = getAppShellLayout;
