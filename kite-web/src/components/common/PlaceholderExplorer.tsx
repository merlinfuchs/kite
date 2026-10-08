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

interface PlaceholderGroup {
  label: string;
  placeholders: {
    label: string;
    value: string;
  }[];
}

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
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const placeholderGroups = useMemo(() => {
    return placeholders.filter((group) => group.placeholders.length > 0);
  }, [placeholders]);

  // Searching looks through every category, otherwise only the picked one is
  // listed. The picked category can disappear, e.g. when a block is removed.
  const searching = search.trim() !== "";
  const activeGroup =
    placeholderGroups.find((group) => group.label === category) ??
    placeholderGroups[0];
  const visibleGroups = searching
    ? placeholderGroups
    : activeGroup
    ? [activeGroup]
    : [];

  function selectCategory(label: string) {
    setCategory(label);
    setSearch("");
    if (listRef.current) listRef.current.scrollTop = 0;
    // Keeps typing and the arrow keys working after clicking a category.
    inputRef.current?.focus();
  }

  // Up and down move through the placeholders, so left and right are free to
  // move through the categories while nothing is typed.
  function onKeyDown(e: KeyboardEvent) {
    if (searching || !activeGroup) return;
    if (e.key !== "ArrowLeft" && e.key !== "ArrowRight") return;

    const index = placeholderGroups.indexOf(activeGroup);
    const next = placeholderGroups[index + (e.key === "ArrowRight" ? 1 : -1)];
    if (next) {
      e.preventDefault();
      selectCategory(next.label);
    }
  }

  return (
    <Popover
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        if (!open) setSearch("");
      }}
      modal
    >
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="w-[500px] max-w-[calc(100vw-1rem)] p-0">
        <Command onKeyDown={onKeyDown}>
          <CommandInput
            ref={inputRef}
            placeholder="Search placeholder..."
            value={search}
            onValueChange={setSearch}
          />
          {tabs && (
            <div className="px-3 pt-3">
              <Tabs value={tab} onValueChange={onTabChange}>
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
          <div className="flex min-h-0">
            {placeholderGroups.length > 0 && (
              <div className="w-[160px] flex-none max-h-[300px] overflow-y-auto border-r p-1 space-y-0.5">
                <div className="px-2 py-1.5 text-xs font-medium text-muted-foreground">
                  Categories
                </div>
                {placeholderGroups.map((group) => (
                  <button
                    type="button"
                    key={group.label}
                    title={group.label}
                    onClick={() => selectCategory(group.label)}
                    className={cn(
                      "flex w-full items-center justify-between gap-2 rounded-sm px-2 py-1.5 text-sm text-left outline-none hover:bg-accent/50 focus-visible:bg-accent/50",
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
            <CommandList ref={listRef} className="flex-auto min-w-0">
              <CommandEmpty>No placeholder found.</CommandEmpty>
              {visibleGroups.map((group) => (
                <CommandGroup heading={group.label} key={group.label}>
                  {group.placeholders.map((placeholder) => (
                    <CommandItem
                      key={placeholder.value}
                      value={placeholder.value}
                      keywords={[placeholder.label, group.label]}
                      onSelect={() => {
                        onSelect(placeholder.value);
                        setOpen(false);
                        setSearch("");
                      }}
                      className="flex flex-col items-start"
                    >
                      <div>{placeholder.label}</div>
                      <div className="text-xs">
                        {!hideBrackets ? (
                          <>
                            <span className="text-muted-foreground mr-1">
                              {"{{"}
                            </span>
                            <span>{placeholder.value}</span>
                            <span className="text-muted-foreground ml-1">
                              {"}}"}
                            </span>
                          </>
                        ) : (
                          <span>{placeholder.value}</span>
                        )}
                      </div>
                    </CommandItem>
                  ))}
                </CommandGroup>
              ))}
            </CommandList>
          </div>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
