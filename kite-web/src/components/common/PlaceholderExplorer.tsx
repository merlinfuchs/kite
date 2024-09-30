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
import { ReactNode, useState } from "react";

interface PlaceholderGroup {
  label: string;
  placeholders: {
    label: string;
    value: string;
  }[];
}

export default function PlaceholderExplorer({
  children,
  onSelect,
  placeholders,
}: {
  children: ReactNode;
  onSelect: (value: string) => void;
  placeholders: PlaceholderGroup[];
}) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="w-[350px] p-0">
        <Command>
          <CommandInput placeholder="Search placeholder..." />
          <CommandList>
            <CommandEmpty>No placeholder found.</CommandEmpty>
            {placeholders.map((group) => (
              <CommandGroup heading={group.label} key={group.label}>
                {group.placeholders.map((placeholder) => (
                  <CommandItem
                    key={placeholder.value}
                    value={placeholder.value}
                    onSelect={(currentValue) => {
                      onSelect(currentValue);
                      setOpen(false);
                    }}
                    className="flex flex-col items-start"
                  >
                    <div>{placeholder.label}</div>
                    <div className="text-xs">
                      <span className="text-muted-foreground mr-1">{"{{"}</span>
                      <span>{placeholder.value}</span>
                      <span className="text-muted-foreground ml-1">{"}}"}</span>
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
