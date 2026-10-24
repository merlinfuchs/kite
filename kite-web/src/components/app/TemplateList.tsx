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
  last,
}: {
  templates: Template[];
  href: (template: Template) => LinkProps["href"];
  added?: (template: Template) => boolean;
  // Shown after the templates, like a card for building your own.
  last?: React.ReactNode;
}) {
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<TemplateCategory | null>(null);

  // Only categories that have a template, so no filter ends up empty.
  const categories = templateCategories.filter((c) =>
    templates.some((t) => t.category === c.value)
  );

  const filtered = templates.filter(
    (t) =>
      (!category || t.category === category) &&
      matchesSearch(search, [
        t.name,
        t.description,
        ...t.commands.map((c) => c.name),
      ])
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
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {filtered.map((template) => (
            <TemplateListEntry
              key={template.id}
              template={template}
              href={href(template)}
              added={added?.(template)}
            />
          ))}
          {last}
        </div>
      ) : (
        <div className="text-sm text-muted-foreground py-12 text-center">
          No templates match your search.
        </div>
      )}
    </div>
  );
}
