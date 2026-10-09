import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import Link from "next/link";
import { ShieldCheckIcon, UploadIcon } from "lucide-react";
import AppLayout from "@/components/app/AppLayout";
import { getAppShellLayout } from "@/components/app/AppShell";
import { TemplateList } from "@/components/app/TemplateList";
import MarketplaceListingDialog from "@/components/marketplace/MarketplaceListingDialog";
import MarketplaceListingGrid from "@/components/marketplace/MarketplaceListingGrid";
import MarketplacePublishDialog from "@/components/marketplace/MarketplacePublishDialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  MarketplaceListingKind,
  MarketplaceListingSort,
  useMarketplaceListingsQuery,
  useMarketplaceMeQuery,
  useMarketplaceMyListingsQuery,
} from "@/lib/api/marketplace";
import { useAppId } from "@/lib/hooks/params";
import env from "@/lib/env/client";

const breadcrumbs = [
  {
    label: "Marketplace",
  },
];

const pageSize = 24;
const tabs = ["browse", "official", "mine"] as const;
type Tab = (typeof tabs)[number];

export default function AppMarketplacePage() {
  const router = useRouter();
  const appId = useAppId();

  const tab: Tab = tabs.includes(router.query.tab as Tab)
    ? (router.query.tab as Tab)
    : "browse";
  const listingId = (router.query.listing as string) || null;

  function setQuery(query: Record<string, string | undefined>) {
    const next = { ...router.query, ...query };
    for (const key of Object.keys(next)) {
      if (next[key] === undefined) delete next[key];
    }
    router.replace({ pathname: router.pathname, query: next }, undefined, {
      shallow: true,
    });
  }

  const me = useMarketplaceMeQuery();
  const isModerator = !!(me.data?.success && me.data.data.is_moderator);

  return (
    <AppLayout title="Marketplace" breadcrumbs={breadcrumbs}>
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold md:text-2xl mb-1">
            Marketplace
          </h1>
          <p className="text-muted-foreground text-sm">
            Import commands, event listeners, message templates and whole
            modules made by the community, or share your own.{" "}
            <a
              href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/marketplace`}
              target="_blank"
              className="text-primary hover:underline"
            >
              Learn More
            </a>
          </p>
        </div>
        <div className="flex gap-2 flex-none">
          {isModerator && (
            <Button variant="outline" asChild>
              <Link
                href={{
                  pathname: "/apps/[appId]/marketplace/moderation",
                  query: { appId },
                }}
              >
                <ShieldCheckIcon className="h-4 w-4 mr-2" />
                Moderation
              </Link>
            </Button>
          )}
          <MarketplacePublishDialog>
            <Button>
              <UploadIcon className="h-4 w-4 mr-2" />
              Publish
            </Button>
          </MarketplacePublishDialog>
        </div>
      </div>
      <Separator className="my-8" />

      <Tabs value={tab} onValueChange={(v) => setQuery({ tab: v })}>
        <TabsList className="mb-6">
          <TabsTrigger value="browse">Community</TabsTrigger>
          <TabsTrigger value="official">Official</TabsTrigger>
          <TabsTrigger value="mine">My Listings</TabsTrigger>
        </TabsList>
        <TabsContent value="browse">
          <BrowseTab onSelect={(id) => setQuery({ listing: id })} />
        </TabsContent>
        <TabsContent value="official">
          <p className="text-muted-foreground text-sm mb-5">
            Ready-made setups maintained by the Kite team. These used to be the
            Templates page.
          </p>
          <TemplateList />
        </TabsContent>
        <TabsContent value="mine">
          <MineTab onSelect={(id) => setQuery({ listing: id })} />
        </TabsContent>
      </Tabs>

      <MarketplaceListingDialog
        listingId={listingId}
        onClose={() => setQuery({ listing: undefined })}
      />
    </AppLayout>
  );
}

function BrowseTab({ onSelect }: { onSelect: (id: string) => void }) {
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [kind, setKind] = useState<MarketplaceListingKind>("");
  const [sort, setSort] = useState<MarketplaceListingSort>("popular");
  const [page, setPage] = useState(0);

  useEffect(() => {
    const timeout = setTimeout(() => {
      setSearch(searchInput.trim());
      setPage(0);
    }, 300);
    return () => clearTimeout(timeout);
  }, [searchInput]);

  const query = useMarketplaceListingsQuery({
    search,
    kind,
    sort,
    limit: pageSize,
    offset: page * pageSize,
  });
  const listings = query.data?.success ? query.data.data : undefined;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row gap-3">
        <Input
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          placeholder="Search listings..."
          maxLength={100}
          className="sm:max-w-sm"
        />
        <Select
          value={kind || "all"}
          onValueChange={(v) => {
            setKind(v === "all" ? "" : (v as MarketplaceListingKind));
            setPage(0);
          }}
        >
          <SelectTrigger className="sm:w-48">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Everything</SelectItem>
            <SelectItem value="command">Commands</SelectItem>
            <SelectItem value="event_listener">Event Listeners</SelectItem>
            <SelectItem value="message">Message Templates</SelectItem>
            <SelectItem value="module">Modules</SelectItem>
          </SelectContent>
        </Select>
        <Select
          value={sort}
          onValueChange={(v) => {
            setSort(v as MarketplaceListingSort);
            setPage(0);
          }}
        >
          <SelectTrigger className="sm:w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="popular">Most imported</SelectItem>
            <SelectItem value="recent">Recently updated</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <MarketplaceListingGrid
        listings={listings}
        loading={query.isLoading}
        emptyTitle="Nothing here yet"
        emptyDescription={
          search
            ? "No listings match your search."
            : "Be the first to publish something!"
        }
        onSelect={onSelect}
      />

      {(page > 0 || (listings?.length ?? 0) >= pageSize) && (
        <div className="flex justify-center gap-2">
          <Button
            variant="outline"
            disabled={page === 0}
            onClick={() => setPage(page - 1)}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            disabled={(listings?.length ?? 0) < pageSize}
            onClick={() => setPage(page + 1)}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}

function MineTab({ onSelect }: { onSelect: (id: string) => void }) {
  const query = useMarketplaceMyListingsQuery();
  const listings = query.data?.success ? query.data.data : undefined;

  return (
    <MarketplaceListingGrid
      listings={listings}
      loading={query.isLoading}
      showStatus
      emptyTitle="You haven't published anything"
      emptyDescription="Use the Publish button to share commands and event listeners from this app."
      onSelect={onSelect}
    />
  );
}

AppMarketplacePage.getLayout = getAppShellLayout;
