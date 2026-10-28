import {
  featuredTemplateIds,
  Template,
  TemplateCategory,
  templateCategories,
} from "@/lib/flow/templates";
import { matchesSearch } from "@/lib/search";
import { cn } from "@/lib/utils";
import { ArrowRightIcon, SearchIcon, SparklesIcon } from "lucide-react";
import Link, { LinkProps } from "next/link";
import { useRouter } from "next/router";
import { useEffect, useRef, useState } from "react";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import TemplateIcon from "./TemplateIcon";
import { FeaturedTemplateEntry, TemplateListEntry } from "./TemplateListEntry";

type Source = "all" | "kite" | "users";
type Sort = "installs" | "newest" | "name";
const sorts: Sort[] = ["installs", "newest", "name"];

const sources: { value: Source; label: string }[] = [
  { value: "all", label: "Anyone" },
  { value: "kite", label: "Kite" },
  { value: "users", label: "Users" },
];

// How many templates the rows show.
const rowSize = 6;

export function TemplateList({
  templates,
  href,
  added,
  buildHref,
  hero,
  rows = true,
  categoryHref,
}: {
  templates: Template[];
  href: (template: Template) => LinkProps["href"];
  added?: (template: Template) => boolean;
  // Where users build their own, offered when nothing matches and after all
  // templates.
  buildHref: LinkProps["href"];
  // The title of a page that's only the marketplace, shown with the search
  // in a band across the page.
  hero?: React.ReactNode;
  // Whether it starts with rows of picks instead of all templates, like on
  // the page of a category.
  rows?: boolean;
  // The page of a category, which the row of a category links to. Without
  // one, it filters by the category.
  categoryHref?: (category: TemplateCategory) => LinkProps["href"];
}) {
  // The filters are in the URL, so links can open the marketplace with them
  // and going back keeps them.
  const router = useRouter();
  const param = (key: string) =>
    typeof router.query[key] === "string" ? router.query[key] : undefined;
  const category =
    templateCategories.find((c) => c.value === param("category"))?.value ??
    null;
  const source = (sources.find((s) => s.value === param("source"))?.value ??
    "all") as Source;
  const sort = (sorts.find((s) => s === param("sort")) ?? "installs") as Sort;
  const setQuery = (values: Record<string, string | null>) => {
    const query = { ...router.query };
    for (const [key, value] of Object.entries(values)) {
      if (value) query[key] = value;
      else delete query[key];
    }
    router.replace({ pathname: router.pathname, query }, undefined, {
      shallow: true,
      scroll: false,
    });
  };
  const setCategory = (value: TemplateCategory | null) =>
    setQuery({ category: value });
  const setSource = (value: Source) =>
    setQuery({ source: value === "all" ? null : value });
  // A sort in the URL shows all templates instead of the rows.
  const setSort = (value: Sort) => setQuery({ sort: value });

  // The search is typed here and goes to the URL a moment later.
  const [search, setSearch] = useState("");
  const urlSearch = param("q") ?? "";
  // Like after going back. Typing a space doesn't change the URL.
  useEffect(() => {
    setSearch((current) =>
      current.trim() === urlSearch ? current : urlSearch
    );
  }, [urlSearch]);
  useEffect(() => {
    if (search.trim() === urlSearch) return;
    const timeout = setTimeout(
      () => setQuery({ q: search.trim() || null }),
      300
    );
    return () => clearTimeout(timeout);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search]);
  // Only the ones the app added, which the Bundles page lists too.
  const [addedOnly, setAddedOnly] = useState(false);
  const allRef = useRef<HTMLDivElement>(null);

  const hasUsers = templates.some((t) => t.author.kind !== "kite");
  // Without filters it shows the rows instead of all templates.
  const browsing =
    rows &&
    !search &&
    !category &&
    source === "all" &&
    !addedOnly &&
    !param("sort");

  const filtered = templates.filter(
    (t) =>
      (!category || t.category === category) &&
      (source === "all" ||
        (source === "kite") === (t.author.kind === "kite")) &&
      (!addedOnly || !!added?.(t)) &&
      matchesSearch(search, [
        t.name,
        t.description,
        t.author.kind === "user" ? t.author.name : undefined,
        ...t.commands.map((c) => c.name),
      ])
  );
  const sorted = sortTemplates(filtered, sort);

  const entry = (t: Template) => (
    <TemplateListEntry
      key={t.id}
      template={t}
      href={href(t)}
      added={added?.(t)}
    />
  );

  const controls = (
    // Without the categories, the toggle fits next to the search.
    <div
      className={cn(
        hero ? "flex flex-col md:flex-row md:items-center gap-3" : "space-y-4"
      )}
    >
      <div className={cn("relative", hero && "flex-1")}>
        <SearchIcon className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search templates, commands like /ban, or creators"
          className="h-12 pl-11 text-base bg-background"
        />
      </div>
      <div className="flex flex-col-reverse md:flex-row gap-3 md:items-center justify-between">
        {/* The marketplace bar has the categories of the public pages. */}
        <div
          className={cn(
            "flex gap-1.5 overflow-x-auto [scrollbar-width:none] md:flex-wrap",
            hero && "hidden"
          )}
        >
          <Chip
            active={category === null}
            onClick={() => setCategory(null)}
            label="All"
            count={templates.length}
          />
          {templateCategories.map((c) => {
            // Only categories that have a template, so no filter ends up
            // empty.
            const count = templates.filter(
              (t) => t.category === c.value
            ).length;
            if (!count) return null;
            return (
              <Chip
                key={c.value}
                active={category === c.value}
                onClick={() => setCategory(c.value)}
                label={c.label}
                count={count}
              />
            );
          })}
          {added && templates.some(added) && (
            <Chip
              active={addedOnly}
              onClick={() => setAddedOnly(!addedOnly)}
              label="Added"
            />
          )}
        </div>
        {hasUsers && (
          <div className="flex flex-none self-start rounded-lg border bg-background p-0.5">
            {sources.map((s) => (
              <button
                key={s.value}
                onClick={() => setSource(s.value)}
                className={cn(
                  "px-3 py-1 text-sm rounded-md text-muted-foreground transition-colors",
                  source === s.value && "bg-muted text-foreground font-medium"
                )}
              >
                {s.label}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );

  // After the rows are gone, from the top of the templates.
  const scrollToAll = () =>
    setTimeout(() => allRef.current?.scrollIntoView({ behavior: "smooth" }));

  const content = browsing ? (
    <Rows
      templates={templates}
      href={href}
      added={added}
      entry={entry}
      showAll={(sort, source) => {
        setQuery({ sort, source: source ?? null });
        scrollToAll();
      }}
      categoryHref={categoryHref}
    />
  ) : (
    <section>
      <div className="flex items-center justify-between gap-4 mb-4">
        <h2 className="text-xl font-semibold">
          {sorted.length} {sorted.length === 1 ? "template" : "templates"}
        </h2>
        <Select value={sort} onValueChange={(v) => setSort(v as Sort)}>
          <SelectTrigger className="w-36 h-9">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="installs">Most added</SelectItem>
            <SelectItem value="newest">Newest</SelectItem>
            <SelectItem value="name">Name</SelectItem>
          </SelectContent>
        </Select>
      </div>
      {sorted.length ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {sorted.map(entry)}
          {!search && <BuildYourOwnCard href={buildHref} />}
        </div>
      ) : (
        <NoMatches search={search} href={buildHref} />
      )}
    </section>
  );

  if (hero) {
    return (
      <>
        <section className="border-b bg-muted/30">
          <div className="container max-w-5xl pt-12 pb-8 md:pt-16 space-y-6">
            {hero}
            {controls}
          </div>
        </section>
        <div ref={allRef} className="container max-w-5xl py-10 scroll-mt-20">
          {content}
        </div>
      </>
    );
  }

  return (
    <div className="space-y-10">
      {controls}
      <div ref={allRef} className="scroll-mt-20">
        {content}
      </div>
    </div>
  );
}

// Rows are picks from the templates: the featured ones, the most added, the
// newest of the users, and the categories.
function Rows({
  templates,
  href,
  added,
  entry,
  showAll,
  categoryHref,
}: {
  templates: Template[];
  href: (template: Template) => LinkProps["href"];
  added?: (template: Template) => boolean;
  entry: (template: Template) => React.ReactNode;
  showAll: (sort: Sort, source?: Source) => void;
  categoryHref?: (category: TemplateCategory) => LinkProps["href"];
}) {
  const featured = featuredTemplateIds.flatMap(
    (id) => templates.find((t) => t.id === id) ?? []
  );
  // Without the featured ones, which are right above.
  const mostAdded = sortTemplates(templates, "installs")
    .filter((t) => t.installs && !featured.includes(t))
    .slice(0, rowSize);
  const users = templates.filter((t) => t.author.kind !== "kite");
  // One per creator, so one who published a few at once doesn't fill it.
  const newest = sortTemplates(users, "newest")
    .filter(
      (t, i, all) => all.findIndex((o) => creatorKey(o) === creatorKey(t)) === i
    )
    .slice(0, rowSize);

  return (
    <div className="space-y-12">
      {!!featured.length && (
        <Row
          title="Featured"
          description="Picked by Kite, from its own templates and the community's."
        >
          <div className={cn(swipe, "md:grid-cols-3")}>
            {featured.map((t) => (
              <FeaturedTemplateEntry
                key={t.id}
                template={t}
                href={href(t)}
                added={added?.(t)}
              />
            ))}
          </div>
        </Row>
      )}
      {/* Without pages for them, like in the app, the chips have them. */}
      {categoryHref && (
        <Row
          title="Browse by category"
          description="Every template, sorted by what it's for."
        >
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {templateCategories.map((c) => {
              const inCategory = sortTemplates(
                templates.filter((t) => t.category === c.value),
                "installs"
              );
              if (!inCategory.length) return null;
              return (
                <CategoryTile
                  key={c.value}
                  category={c.value}
                  templates={inCategory}
                  href={categoryHref(c.value)}
                />
              );
            })}
          </div>
        </Row>
      )}
      {!!mostAdded.length && (
        <Row
          title="Most added"
          description="The templates the most bots use."
          action={{ label: "See all", onClick: () => showAll("installs") }}
        >
          <div className={cn(swipe, "sm:grid-cols-2 lg:grid-cols-3")}>
            {mostAdded.map(entry)}
          </div>
        </Row>
      )}
      {!!newest.length && (
        <Row
          title="New from the community"
          description="Made by users of Kite, who are responsible for them. Kite reviews them before they're listed."
          action={{
            label: `See all ${users.length}`,
            onClick: () => showAll("newest", "users"),
          }}
        >
          <div className={cn(swipe, "sm:grid-cols-2 lg:grid-cols-3")}>
            {newest.map(entry)}
          </div>
        </Row>
      )}
    </div>
  );
}

// CategoryTile is a category with the icons and names of its most added
// templates, sorted by installs.
function CategoryTile({
  category,
  templates,
  href,
}: {
  category: TemplateCategory;
  templates: Template[];
  href: LinkProps["href"];
}) {
  const c = templateCategories.find((c) => c.value === category)!;
  const top = templates.slice(0, 3);

  return (
    <Link
      href={href}
      className={cn(
        "rounded-xl border bg-gradient-to-br to-transparent p-5 flex flex-col gap-4 transition-colors hover:border-primary/60",
        c.banner
      )}
    >
      <div>
        <div className="flex items-baseline justify-between gap-3">
          <div className="text-lg font-semibold">{c.label}</div>
          <div className="text-sm text-muted-foreground flex-none">
            {templates.length}{" "}
            {templates.length === 1 ? "template" : "templates"}
          </div>
        </div>
        <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
          {c.description}
        </p>
      </div>
      <div className="mt-auto flex items-center gap-3 min-w-0">
        <div className="flex -space-x-2 flex-none">
          {top.map((t) => (
            <div
              key={t.id}
              className="h-8 w-8 rounded-lg border bg-background flex items-center justify-center"
            >
              <TemplateIcon name={t.icon} className={cn("h-4 w-4", c.icon)} />
            </div>
          ))}
        </div>
        <div className="text-xs text-muted-foreground truncate">
          {top.map((t) => t.name).join(", ")}
        </div>
      </div>
    </Link>
  );
}

// swipe makes a row scroll sideways on phones instead of stacking its cards.
const swipe =
  "flex gap-4 overflow-x-auto snap-x snap-mandatory [scrollbar-width:none] [&>*]:w-[80%] [&>*]:flex-none [&>*]:snap-start sm:grid sm:overflow-visible sm:[&>*]:w-auto";

function Row({
  title,
  description,
  action,
  children,
}: {
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
  children: React.ReactNode;
}) {
  return (
    <section>
      <div className="flex items-end justify-between gap-4 mb-4">
        <div>
          <h2 className="text-xl font-semibold">{title}</h2>
          {description && (
            <p className="text-sm text-muted-foreground mt-1">{description}</p>
          )}
        </div>
        {action && (
          <button
            onClick={action.onClick}
            className="flex-none flex items-center gap-1 text-sm text-primary hover:underline"
          >
            {action.label}
            <ArrowRightIcon className="h-4 w-4" />
          </button>
        )}
      </div>
      {children}
    </section>
  );
}

function Chip({
  active,
  onClick,
  label,
  count,
}: {
  active: boolean;
  onClick: () => void;
  label: string;
  count?: number;
}) {
  return (
    <Button
      size="sm"
      variant={active ? "secondary" : "ghost"}
      className="rounded-full flex-none"
      onClick={onClick}
    >
      {label}
      {count !== undefined && (
        <span className="ml-1 text-muted-foreground font-normal">{count}</span>
      )}
    </Button>
  );
}

function sortTemplates(templates: Template[], sort: Sort) {
  return [...templates].sort((a, b) => {
    if (sort === "installs") return (b.installs ?? 0) - (a.installs ?? 0);
    if (sort === "newest") {
      return (b.createdAt ?? "").localeCompare(a.createdAt ?? "");
    }
    return a.name.localeCompare(b.name);
  });
}

function creatorKey(template: Template) {
  return template.author.kind === "user"
    ? template.author.discordId
    : template.author.kind;
}

// BuildYourOwnCard has the shape of a template card, with what the user
// would type for the AI where the cards have a command.
function BuildYourOwnCard({ href }: { href: LinkProps["href"] }) {
  return (
    <Link
      href={href}
      className="h-full flex flex-col gap-3 rounded-xl border border-dashed p-4 transition-colors hover:border-primary/60 hover:bg-muted/30"
    >
      <div className="flex items-center gap-3">
        <div className="h-10 w-10 flex-none rounded-lg bg-primary/10 flex items-center justify-center">
          <SparklesIcon className="h-5 w-5 text-primary" />
        </div>
        <div>
          <div className="font-semibold">Build your own</div>
          <div className="text-xs text-muted-foreground">
            With Kite&apos;s AI
          </div>
        </div>
      </div>
      <div className="text-sm text-muted-foreground line-clamp-2">
        Not here? Describe the command you need and the AI builds it for you.
      </div>
      <div className="mt-auto rounded-md bg-muted/60 px-3 py-2 text-sm flex items-center gap-2 min-w-0">
        <span className="text-xs text-muted-foreground truncate">
          Describe your command…
        </span>
        <ArrowRightIcon className="ml-auto h-3.5 w-3.5 flex-none text-primary" />
      </div>
    </Link>
  );
}

// NoMatches offers to build what the user searched for, so the search isn't
// a dead end.
function NoMatches({
  search,
  href,
}: {
  search: string;
  href: LinkProps["href"];
}) {
  return (
    <div className="rounded-xl border border-dashed py-12 px-6 text-center">
      <div className="h-12 w-12 mx-auto rounded-xl bg-primary/15 flex items-center justify-center mb-4">
        <SparklesIcon className="h-6 w-6 text-primary" />
      </div>
      <div className="font-semibold">
        {search.trim()
          ? `No templates for “${search.trim()}”`
          : "No templates match"}
      </div>
      <p className="text-sm text-muted-foreground mt-1 max-w-md mx-auto">
        Describe what you need and Kite&apos;s AI builds it for you, or connect
        the blocks yourself in the editor.
      </p>
      <Button className="mt-5" asChild>
        <Link href={href}>Build it with AI</Link>
      </Button>
    </div>
  );
}
