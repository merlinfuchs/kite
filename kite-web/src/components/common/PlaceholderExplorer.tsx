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
import { ReactNode, useMemo, useState } from "react";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";
import { cn } from "@/lib/utils";

interface PlaceholderGroup {
  label: string;
  placeholders: {
    label: string;
    value: string;
    // Shown next to the label in a muted color.
    description?: string;
    // The value is inserted as is instead of being wrapped in {{ }}.
    raw?: boolean;
    // Extra search terms, the value is always searchable.
    keywords?: string[];
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
  onOpenChange,
  header,
  hideBrackets,
}: {
  children: ReactNode;
  onSelect: (value: string, raw?: boolean) => void;
  placeholders: PlaceholderGroup[];
  tab?: string;
  tabs?: {
    label: string;
    value: string;
  }[];
  onTabChange?: (value: string) => void;
  onOpenChange?: (open: boolean) => void;
  // Shown above the placeholders, e.g. for inputs that change them.
  header?: ReactNode;
  hideBrackets?: boolean;
}) {
  const [open, setOpen] = useState(false);

  const placeholderGroups = useMemo(() => {
    return placeholders.filter((group) => group.placeholders.length > 0);
  }, [placeholders]);

  return (
    <Popover
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        onOpenChange?.(open);
      }}
      modal
    >
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="w-[350px] p-0">
        <Command>
          <CommandInput placeholder="Search placeholder..." />
          <CommandList>
            <CommandEmpty>No placeholder found.</CommandEmpty>
            {tabs && (
              <CommandGroup className="px-2 pt-2">
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
              </CommandGroup>
            )}
            {header && (
              // Keys typed into the header must not move the list selection.
              <div onKeyDown={(e) => e.stopPropagation()}>{header}</div>
            )}
            {placeholderGroups.map((group) => (
              <CommandGroup heading={group.label} key={group.label}>
                {group.placeholders.map((placeholder) => (
                  <CommandItem
                    key={placeholder.value}
                    value={placeholder.value}
                    keywords={placeholder.keywords}
                    onSelect={() => {
                      onSelect(placeholder.value, placeholder.raw);
                      setOpen(false);
                    }}
                    className="flex flex-col items-start"
                  >
                    <div>
                      {placeholder.label}
                      {placeholder.description && (
                        <span className="text-muted-foreground ml-2 text-xs">
                          {placeholder.description}
                        </span>
                      )}
                    </div>
                    <div className="text-xs">
                      {!hideBrackets && !placeholder.raw ? (
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
        </Command>
      </PopoverContent>
    </Popover>
  );
}
