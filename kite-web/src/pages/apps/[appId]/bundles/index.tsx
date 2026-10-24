import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import BundleDialog from "@/components/app/BundleDialog";
import BundleExportDialog from "@/components/app/BundleExportDialog";
import BundleImportDialog from "@/components/app/BundleImportDialog";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { useBundleDeleteMutation } from "@/lib/api/mutations";
import env from "@/lib/env/client";
import { useApps, useUserBundles } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { UserBundle } from "@/lib/types/wire.gen";
import {
  ImportIcon,
  PackageIcon,
  PlusIcon,
  ShareIcon,
  Trash2Icon,
} from "lucide-react";
import Link from "next/link";
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
    </AppLayout>
  );
}

function BundleEntry({ bundle }: { bundle: UserBundle }) {
  const appId = useAppId();
  const apps = useApps();
  const deleteMutation = useBundleDeleteMutation();

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
          </div>
        </div>
      </Link>
      <div className="flex gap-2 flex-none">
        {fromHere && (
          <BundleDialog bundle={bundle}>
            <Button variant="outline">Publish update</Button>
          </BundleDialog>
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

AppBundlesPage.getLayout = getAppShellLayout;
