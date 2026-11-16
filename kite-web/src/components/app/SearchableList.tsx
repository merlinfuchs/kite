import { ReactNode, useState } from "react";
import {
  ChevronDownIcon,
  FolderIcon,
  PackageIcon,
  SearchIcon,
  XIcon,
} from "lucide-react";
import { Input } from "../ui/input";
import { matchesSearch } from "@/lib/search";
import { ItemGroup, sortByGroup } from "@/lib/flow/itemGroups";
import { foldedGroupKey, useFoldedGroupsStore } from "@/lib/state/foldedGroups";
import { cn } from "@/lib/utils";

// Short lists are quicker to scan than to search.
const minSearchItems = 7;

// Returns a fragment so the items stay direct children of the list's
// AutoAnimate container.
export default function SearchableList<T>({
  items,
  fields,
  renderItem,
  placeholder,
  actions,
  status,
  groups,
}: {
  items: (T | undefined)[];
  fields: (item: T) => (string | null | undefined)[];
  renderItem: (item: T) => ReactNode;
  placeholder: string;
  actions: ReactNode;
  // Shown above the search and actions, like how much of the plan's limit is
  // used.
  status?: ReactNode;
  // Lists the items under a header per group, after the ones without one.
  // Folded groups are remembered per list.
  groups?: {
    list: string;
    of: (item: T) => ItemGroup | undefined;
    actions?: (group: ItemGroup) => ReactNode;
  };
}) {
  const [search, setSearch] = useState("");
  const folded = useFoldedGroupsStore((s) => s.folded);
  const toggleFolded = useFoldedGroupsStore((s) => s.toggle);

  const searchable = items.length >= minSearchItems;
  const query = searchable ? search : "";
  const filtered = items.filter(
    (item): item is T => !!item && matchesSearch(query, fields(item))
  );

  return (
    <>
      {status}
      <div className="flex flex-col gap-3 md:flex-row md:items-center">
        <div className="md:ml-auto">
          <ListActions>{actions}</ListActions>
        </div>
        {searchable && (
          <div className="relative md:order-first md:flex-1 md:max-w-sm">
            <SearchIcon className="absolute size-4 left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Escape") setSearch("");
              }}
              placeholder={placeholder}
              aria-label={placeholder}
              className="pl-9 pr-9"
            />
            {search && (
              <button
                type="button"
                onClick={() => setSearch("")}
                aria-label="Clear search"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                <XIcon className="size-4" />
              </button>
            )}
          </div>
        )}
      </div>
      {filtered.length === 0 ? (
        <div className="text-sm text-muted-foreground">
          No results for &quot;{query.trim()}&quot;.{" "}
          <button
            type="button"
            onClick={() => setSearch("")}
            className="underline hover:text-foreground"
          >
            Clear search
          </button>
        </div>
      ) : groups ? (
        renderGroups()
      ) : (
        filtered.map(renderItem)
      )}
    </>
  );

  // The headers are siblings of the items, so the items stay direct children
  // of the list's AutoAnimate container.
  function renderGroups() {
    const of = groups!.of;
    const sorted = sortByGroup(filtered, (item) => of(item)?.name);
    const nodes: ReactNode[] = [];
    let current: string | undefined;
    for (const item of sorted) {
      const group = of(item);
      if (group && group.name !== current) {
        const key = foldedGroupKey(groups!.list, group.name);
        // A search shows the matches in folded groups too.
        const isFolded = !query && !!folded[key];
        nodes.push(
          <GroupHeader
            key={"group:" + group.name}
            group={group}
            count={sorted.filter((i) => of(i)?.name === group.name).length}
            folded={isFolded}
            onToggle={query ? undefined : () => toggleFolded(key)}
            actions={groups!.actions?.(group)}
          />
        );
      }
      current = group?.name;
      if (group && !query && folded[foldedGroupKey(groups!.list, group.name)]) {
        continue;
      }
      nodes.push(renderItem(item));
    }
    return nodes;
  }
}

function GroupHeader({
  group,
  count,
  folded,
  onToggle,
  actions,
}: {
  group: ItemGroup;
  count: number;
  folded: boolean;
  onToggle?: () => void;
  actions?: ReactNode;
}) {
  const Icon = group.template ? PackageIcon : FolderIcon;

  return (
    <div className="flex items-center gap-2 border-b pb-1 pt-2 -mb-2">
      <button
        type="button"
        onClick={onToggle}
        disabled={!onToggle}
        aria-expanded={!folded}
        className="flex flex-1 min-w-0 items-center gap-2 py-1 text-left disabled:cursor-default"
      >
        <ChevronDownIcon
          className={cn(
            "size-4 flex-none text-muted-foreground transition-transform",
            folded && "-rotate-90"
          )}
        />
        <Icon className="size-4 flex-none text-muted-foreground" />
        <span className="font-medium truncate">{group.name}</span>
        <span className="text-sm text-muted-foreground">{count}</span>
      </button>
      {actions}
    </div>
  );
}

// The create and import buttons of a list: full width on mobile, in a row
// otherwise.
export function ListActions({ children }: { children: ReactNode }) {
  return <div className="flex flex-col gap-3 md:flex-row">{children}</div>;
}
