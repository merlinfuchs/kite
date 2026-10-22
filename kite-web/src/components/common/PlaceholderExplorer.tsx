import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { KeyboardEvent, ReactNode, useMemo, useRef, useState } from "react";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";
import { cn } from "@/lib/utils";
import {
  flattenPlaceholders,
  Placeholder,
  PlaceholderGroup,
} from "@/lib/flow/placeholders";
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import { defaultFilter } from "cmdk";

const backValue = "$back";

const tabGridCols: Record<number, string> = {
  1: "grid-cols-1",
  2: "grid-cols-2",
  3: "grid-cols-3",
};

export default function PlaceholderExplorer({
  children,
  onSelect,
  placeholders,
  tab,
  tabs,
  onTabChange,
  hideBrackets,
}: {
  children: ReactNode;
  onSelect: (value: string) => void;
  placeholders: PlaceholderGroup[];
  tab?: string;
  tabs?: {
    label: string;
    value: string;
  }[];
  onTabChange?: (value: string) => void;
  hideBrackets?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState<string>();
  // The highlighted placeholder, which Enter picks.
  const [selected, setSelected] = useState("");
  // The values whose fields are shown, innermost last.
  const [path, setPath] = useState<Placeholder[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const categoriesRef = useRef<HTMLDivElement>(null);

  const placeholderGroups = useMemo(() => {
    return placeholders.filter((group) => group.placeholders.length > 0);
  }, [placeholders]);

  // Searching looks through every category, otherwise only the picked one is
  // listed. The picked category can disappear, e.g. when a block is removed.
  const searching = search.trim() !== "";
  const activeGroup =
    placeholderGroups.find((group) => group.label === category) ??
    placeholderGroups[0];
  const openValue = path.at(-1);
  // Searching also finds the fields of values, like User Top Role Name.
  const flatGroups = useMemo(
    () =>
      searching
        ? placeholderGroups.map((group) => ({
            ...group,
            placeholders: flattenPlaceholders(group.placeholders),
          }))
        : [],
    [searching, placeholderGroups]
  );
  // cmdk sorts the items in a group, but doesn't move the groups, so the
  // group with the best match goes first here.
  const searchGroups = useMemo(() => {
    const score = (group: PlaceholderGroup) =>
      Math.max(
        0,
        ...group.placeholders.map((p) =>
          defaultFilter(p.value, search, [p.label, group.label])
        )
      );
    return flatGroups
      .map((group) => ({ group, score: score(group) }))
      .sort((a, b) => b.score - a.score)
      .map(({ group }) => group);
  }, [flatGroups, search]);
  const visibleGroups = searching
    ? searchGroups
    : openValue
    ? [
        {
          label: path.map((p) => p.label).join(" › "),
          placeholders: openValue.fields ?? [],
        },
      ]
    : activeGroup
    ? [activeGroup]
    : [];

  const highlighted = visibleGroups
    .flatMap((g) => g.placeholders)
    .find((p) => p.value === selected);

  // The fields of an opened value leave out its value, like .name of
  // user.top_role.name, as it's the same for all of them.
  function shortValue(value: string) {
    if (!openValue?.value) return value;
    const rest = value.slice(openValue.value.length);
    return value.startsWith(openValue.value) && /^[.[]/.test(rest)
      ? rest
      : value;
  }

  function resetList(selected: string) {
    setSelected(selected);
    setSearch("");
    if (listRef.current) listRef.current.scrollTop = 0;
    inputRef.current?.focus();
  }

  function showFields(placeholder: Placeholder, group?: PlaceholderGroup) {
    // A search result opens in its category, which Back returns to.
    if (searching && group) {
      setCategory(group.label);
      setPath([placeholder]);
    } else {
      setPath([...path, placeholder]);
    }
    resetList(placeholder.fields?.[0]?.value ?? "");
  }

  // Only placeholders without fields are inserted. A value with fields opens
  // them instead, as one of them says what it is, like user.id instead of
  // user, which is also the ID.
  function pick(placeholder: Placeholder, group: PlaceholderGroup) {
    if (placeholder.fields?.length) {
      showFields(placeholder, group);
      return;
    }
    onSelect(placeholder.value);
    onOpenChange(false);
  }

  function goBack() {
    if (!openValue) return;
    setPath(path.slice(0, -1));
    // A value opened from search isn't listed where Back leads to.
    const parent = path.at(-2);
    const list = parent ? parent.fields ?? [] : activeGroup?.placeholders ?? [];
    resetList(
      list.some((p) => p.value === openValue.value)
        ? openValue.value
        : list[0]?.value ?? ""
    );
  }

  function selectCategory(group: PlaceholderGroup) {
    setPath([]);
    setCategory(group.label);
    // cmdk would highlight nothing after the list changed. Focusing the input
    // keeps typing and the arrow keys working after clicking a category.
    resetList(group.placeholders[0]?.value ?? "");
    // The arrow keys can pick a category that is scrolled out of view.
    requestAnimationFrame(() =>
      categoriesRef.current
        ?.querySelector('[aria-pressed="true"]')
        ?.scrollIntoView({ block: "nearest", inline: "nearest" })
    );
  }

  function onOpenChange(open: boolean) {
    setOpen(open);
    if (!open) {
      setSearch("");
      setSelected("");
      setPath([]);
    }
  }

  // Up and down move through the placeholders, so left and right are free to
  // open and close values with fields, or else move through the categories,
  // while nothing is typed.
  function onKeyDown(e: KeyboardEvent) {
    if (searching || !activeGroup) return;
    if (e.key !== "ArrowLeft" && e.key !== "ArrowRight") return;

    if (e.key === "ArrowRight" && highlighted?.fields?.length) {
      e.preventDefault();
      showFields(highlighted);
      return;
    }
    if (openValue) {
      if (e.key === "ArrowLeft") {
        e.preventDefault();
        goBack();
      }
      return;
    }

    const index = placeholderGroups.indexOf(activeGroup);
    const next = placeholderGroups[index + (e.key === "ArrowRight" ? 1 : -1)];
    if (next) {
      e.preventDefault();
      selectCategory(next);
    }
  }

  return (
    <Popover open={open} onOpenChange={onOpenChange} modal>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      {/* Only the list shrinks when the screen has less room, like on
          phones, so the search stays visible. */}
      <PopoverContent
        className="flex w-[500px] max-w-[calc(100vw-1rem)] max-h-[var(--radix-popover-content-available-height)] flex-col p-0"
        collisionPadding={8}
      >
        <Command
          className="min-h-0"
          value={selected}
          onValueChange={setSelected}
          onKeyDown={onKeyDown}
        >
          <CommandInput
            ref={inputRef}
            placeholder="Search placeholder..."
            value={search}
            onValueChange={(value) => {
              setSearch(value);
              // The best match is first, which could be scrolled out of view.
              if (listRef.current) listRef.current.scrollTop = 0;
            }}
          />
          {tabs && (
            <div className="px-3 pt-3">
              <Tabs
                value={tab}
                onValueChange={(value) => {
                  setPath([]);
                  onTabChange?.(value);
                }}
              >
                <TabsList
                  className={cn(
                    "w-full grid h-8 py-0",
                    tabGridCols[tabs.length]
                  )}
                >
                  {tabs?.map((tab) => (
                    <TabsTrigger
                      value={tab.value}
                      className="py-0.5 font-normal"
                      key={tab.value}
                    >
                      {tab.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>
            </div>
          )}
          {/* On small screens the categories are a row above the list, so
              the list keeps the full width. */}
          <div className="flex flex-auto flex-col sm:flex-row min-h-0">
            {placeholderGroups.length > 0 && (
              <div
                ref={categoriesRef}
                role="group"
                aria-label="Categories"
                className="flex sm:block flex-none gap-1 overflow-x-auto sm:overflow-x-hidden sm:overflow-y-auto sm:w-[160px] sm:max-h-[300px] border-b sm:border-b-0 sm:border-r p-1 sm:space-y-0.5"
              >
                <div className="hidden sm:block px-2 py-1.5 text-xs font-medium text-muted-foreground">
                  Categories
                </div>
                {placeholderGroups.map((group) => (
                  <button
                    type="button"
                    key={group.label}
                    title={group.label}
                    aria-pressed={!searching && group === activeGroup}
                    onClick={() => selectCategory(group)}
                    // cmdk would pick the highlighted placeholder instead.
                    onKeyDown={(e) => e.key === "Enter" && e.stopPropagation()}
                    className={cn(
                      "flex flex-none sm:w-full items-center justify-between gap-2 rounded-sm px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none hover:bg-accent/50 focus-visible:bg-accent/50",
                      !searching &&
                        group === activeGroup &&
                        "bg-accent text-accent-foreground hover:bg-accent"
                    )}
                  >
                    <span className="truncate">{group.label}</span>
                    <span className="text-xs text-muted-foreground flex-none">
                      {group.placeholders.length}
                    </span>
                  </button>
                ))}
              </div>
            )}
            <CommandList ref={listRef} className="flex-auto min-w-0 min-h-0">
              <CommandEmpty>No placeholder found.</CommandEmpty>
              {visibleGroups.map((group) => (
                <CommandGroup heading={group.label} key={group.label}>
                  {!searching && openValue && (
                    <CommandItem
                      value={backValue}
                      onSelect={goBack}
                      className="text-muted-foreground"
                    >
                      <ChevronLeftIcon className="h-4 w-4 mr-1" />
                      Back
                    </CommandItem>
                  )}
                  {group.placeholders.map((placeholder) => (
                    <CommandItem
                      key={placeholder.value}
                      value={placeholder.value}
                      keywords={[placeholder.label, group.label]}
                      onSelect={() => pick(placeholder, group)}
                      className="flex items-center gap-3"
                    >
                      <span className="flex-auto truncate">
                        {placeholder.label}
                      </span>
                      {/* The labels of search results are long, like User
                          Top Role Name, and the footer shows the value. */}
                      {!searching && (
                        <span className="max-w-[40%] truncate font-mono text-xs text-muted-foreground">
                          {shortValue(placeholder.value)}
                        </span>
                      )}
                      {!!placeholder.fields?.length && (
                        <ChevronRightIcon className="-mr-1 h-4 w-4 flex-none text-muted-foreground" />
                      )}
                    </CommandItem>
                  ))}
                </CommandGroup>
              ))}
            </CommandList>
          </div>
          {highlighted && (
            <div className="border-t px-3 py-2 text-xs">
              {highlighted.fields?.length ? (
                <div className="text-muted-foreground">
                  {fieldsSummary(highlighted.fields)}
                </div>
              ) : (
                <div className="font-mono break-all">
                  {hideBrackets
                    ? highlighted.value
                    : `{{${highlighted.value}}}`}
                </div>
              )}
              {highlighted.description && (
                <div className="mt-1 text-muted-foreground">
                  {highlighted.description}
                </div>
              )}
            </div>
          )}
        </Command>
      </PopoverContent>
    </Popover>
  );
}

// fieldsSummary names the first fields of a value, like "25 fields, like ID,
// Username and Display Name".
function fieldsSummary(fields: Placeholder[]) {
  const names = fields.slice(0, 3).map((f) => f.label);
  const list =
    names.length > 1
      ? `${names.slice(0, -1).join(", ")} and ${names.at(-1)}`
      : names[0];
  return fields.length > names.length
    ? `${fields.length} fields, like ${list}`
    : `Fields: ${list}`;
}
