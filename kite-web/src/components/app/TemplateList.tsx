import {
  Template,
  TemplateCategory,
  templateCategories,
} from "@/lib/flow/templates";
import { matchesSearch } from "@/lib/search";
import { SearchIcon } from "lucide-react";
import { LinkProps } from "next/link";
import { useState } from "react";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { TemplateListEntry } from "./TemplateListEntry";

export function TemplateList({
  templates,
  href,
  added,
  installs,
  last,
}: {
  templates: Template[];
  href: (template: Template) => LinkProps["href"];
  added?: (template: Template) => boolean;
  installs?: (template: Template) => number | undefined;
  // Shown after the templates, like a card for building your own.
  last?: React.ReactNode;
}) {
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<TemplateCategory | null>(null);
  // Only the ones the app added, which the Bundles page lists too.
  const [addedOnly, setAddedOnly] = useState(false);

  // Only categories that have a template, so no filter ends up empty.
  const categories = templateCategories.filter((c) =>
    templates.some((t) => t.category === c.value)
  );

  const filtered = templates.filter(
    (t) =>
      (!category || t.category === category) &&
      (!addedOnly || !!added?.(t)) &&
      matchesSearch(search, [
        t.name,
        t.description,
        ...t.commands.map((c) => c.name),
      ])
  );

  // Official ones are apart from the ones users made, which show who made
  // them.
  const hasCommunity = templates.some((t) => t.author.kind !== "kite");
  const official = filtered.filter((t) => t.author.kind === "kite");
  const community = filtered.filter((t) => t.author.kind !== "kite");
  const grid = (list: Template[], after?: React.ReactNode) => (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {list.map((template) => (
        <TemplateListEntry
          key={template.id}
          template={template}
          href={href(template)}
          added={added?.(template)}
          installs={installs?.(template)}
        />
      ))}
      {after}
    </div>
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col-reverse sm:flex-row gap-3 sm:items-center justify-between">
        <div className="flex flex-wrap gap-1.5">
          <Button
            size="sm"
            variant={category === null ? "secondary" : "ghost"}
            className="rounded-full"
            onClick={() => setCategory(null)}
          >
            All
          </Button>
          {categories.map((c) => (
            <Button
              key={c.value}
              size="sm"
              variant={category === c.value ? "secondary" : "ghost"}
              className="rounded-full"
              onClick={() => setCategory(c.value)}
            >
              {c.label}
            </Button>
          ))}
          {added && templates.some(added) && (
            <Button
              size="sm"
              variant={addedOnly ? "secondary" : "ghost"}
              className="rounded-full"
              onClick={() => setAddedOnly(!addedOnly)}
            >
              Added
            </Button>
          )}
        </div>
        <div className="relative sm:w-64">
          <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search templates"
            className="pl-9"
          />
        </div>
      </div>

      {filtered.length ? (
        <div className="space-y-10">
          {!!official.length && (
            <section>
              {hasCommunity && (
                <h2 className="font-semibold mb-4">Made by Kite</h2>
              )}
              {grid(official, last)}
            </section>
          )}
          {!!community.length && (
            <section>
              <h2 className="font-semibold mb-1">From the community</h2>
              <p className="text-sm text-muted-foreground mb-4">
                Made by users of Kite, who are responsible for them. Kite
                reviews them before they&apos;re listed.
              </p>
              {grid(community)}
            </section>
          )}
        </div>
      ) : (
        <div className="text-sm text-muted-foreground py-12 text-center">
          No templates match your search.
        </div>
      )}
    </div>
  );
}
