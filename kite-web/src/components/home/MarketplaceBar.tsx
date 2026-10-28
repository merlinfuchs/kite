import { Input } from "@/components/ui/input";
import {
  categoryPath,
  TemplateCategory,
  templateCategories,
} from "@/lib/flow/templates";
import { cn } from "@/lib/utils";
import { SearchIcon, StoreIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useState } from "react";

// MarketplaceBar is on every page of the public marketplace, so they feel
// like one place: a link to it, its categories and a search that opens it.
// The marketplace itself has its own search.
export default function MarketplaceBar({
  category,
  search = true,
}: {
  // The one to highlight, like the category of a template, or null for All.
  category?: TemplateCategory | null;
  search?: boolean;
}) {
  const router = useRouter();
  const [query, setQuery] = useState("");

  const links = [
    { value: null, label: "All", href: "/marketplace" },
    ...templateCategories.map((c) => ({
      value: c.value,
      label: c.label,
      href: categoryPath(c.value),
    })),
  ];

  return (
    <div className="border-b bg-muted/30">
      <div className="container max-w-5xl flex items-center gap-4 h-12">
        <Link
          href="/marketplace"
          className="flex-none flex items-center gap-2 font-semibold"
        >
          <StoreIcon className="h-4 w-4 text-primary" />
          Marketplace
        </Link>
        <div className="h-5 w-px bg-border flex-none" />
        <nav className="flex-1 min-w-0 flex gap-1 overflow-x-auto [scrollbar-width:none]">
          {links.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className={cn(
                "flex-none px-2.5 py-1 rounded-md text-sm transition-colors",
                category === link.value
                  ? "bg-muted text-foreground font-medium"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              {link.label}
            </Link>
          ))}
        </nav>
        {search && (
          <form
            className="relative hidden md:block w-56 flex-none"
            onSubmit={(e) => {
              e.preventDefault();
              router.push({
                pathname: "/marketplace",
                query: query.trim() ? { q: query.trim() } : {},
              });
            }}
          >
            <SearchIcon className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search templates"
              className="h-8 pl-8 text-sm bg-background"
            />
          </form>
        )}
      </div>
    </div>
  );
}
