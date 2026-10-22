import { ReactNode, useState } from "react";
import { SearchIcon, XIcon } from "lucide-react";
import { Input } from "../ui/input";
import { matchesSearch } from "@/lib/search";

// Returns a fragment so the items stay direct children of the list's
// AutoAnimate container.
export default function SearchableList<T>({
  items,
  fields,
  renderItem,
  placeholder,
}: {
  items: (T | undefined)[];
  fields: (item: T) => (string | null | undefined)[];
  renderItem: (item: T) => ReactNode;
  placeholder: string;
}) {
  const [search, setSearch] = useState("");

  const filtered = items.filter(
    (item): item is T => !!item && matchesSearch(search, fields(item))
  );

  return (
    <>
      <div className="relative">
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
      {filtered.length === 0 ? (
        <div className="flex items-center justify-center rounded-lg border border-dashed px-4 py-12 text-center text-sm text-muted-foreground">
          No results for &quot;{search.trim()}&quot;
        </div>
      ) : (
        filtered.map(renderItem)
      )}
    </>
  );
}
