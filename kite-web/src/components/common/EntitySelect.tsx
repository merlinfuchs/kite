import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";
import { Button } from "../ui/button";
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "../ui/command";
import { cn } from "@/lib/utils";
import { useState } from "react";

// A searchable select of Discord servers, channels, roles or similar items.
export default function EntitySelect({
  items,
  value,
  onChange,
  placeholder,
  searchPlaceholder,
  emptyText,
  wide,
}: {
  items: { id: string; name: string; description?: string }[] | undefined;
  value: string | null;
  onChange: (value: string | null) => void;
  placeholder: string;
  searchPlaceholder: string;
  emptyText: string;
  // Makes the list as wide as the button, e.g. for longer descriptions.
  // Also lets the list scroll when the select is inside a dialog.
  wide?: boolean;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen} modal={wide}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between truncate flex"
        >
          <div className="truncate">
            {value ? items?.find((i) => i.id === value)?.name : placeholder}
          </div>
          <ChevronsUpDownIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className={cn(
          "p-0",
          wide ? "w-[var(--radix-popover-trigger-width)]" : "w-[200px]"
        )}
      >
        <Command>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            <CommandEmpty>{emptyText}</CommandEmpty>
            <CommandGroup>
              {items?.map((item) => (
                <CommandItem
                  key={item.id}
                  value={item.id}
                  keywords={[item.name, item.description ?? ""]}
                  onSelect={(currentValue) => {
                    onChange(currentValue);
                    setOpen(false);
                  }}
                >
                  <CheckIcon
                    className={cn(
                      "mr-2 h-4 w-4",
                      value === item.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {item.description ? (
                    <div className="min-w-0">
                      <div>{item.name}</div>
                      <div className="text-xs text-muted-foreground font-mono truncate">
                        {item.description}
                      </div>
                    </div>
                  ) : (
                    item.name
                  )}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
