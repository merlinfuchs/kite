import { ReactNode, useState } from "react";
import { SearchIcon, XIcon } from "lucide-react";
import { Input } from "../ui/input";
import { matchesSearch } from "@/lib/search";

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
}: {
  items: (T | undefined)[];
  fields: (item: T) => (string | null | undefined)[];
  renderItem: (item: T) => ReactNode;
  placeholder: string;
  actions: ReactNode;
}) {
  const [search, setSearch] = useState("");

  const searchable = items.length >= minSearchItems;
  const query = searchable ? search : "";
  const filtered = items.filter(
    (item): item is T => !!item && matchesSearch(query, fields(item))
  );

  return (
    <>
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
      ) : (
        filtered.map(renderItem)
      )}
    </>
  );
}

// The create and import buttons of a list: full width on mobile, in a row
// otherwise.
export function ListActions({ children }: { children: ReactNode }) {
  return <div className="flex flex-col gap-3 md:flex-row">{children}</div>;
}
