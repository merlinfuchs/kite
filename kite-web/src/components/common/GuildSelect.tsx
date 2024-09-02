import { useAppStateGuilds } from "@/lib/hooks/api";
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
import { useState } from "react";
import { cn } from "@/lib/utils";

export default function GuildSelect({
  value,
  onChange,
}: {
  value: string | null;
  onChange: (value: string | null) => void;
}) {
  const guilds = useAppStateGuilds();

  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between truncate"
        >
          <div className="truncate">
            {value
              ? guilds?.find((g) => g!.id === value)?.name
              : "Select server..."}
          </div>
          <ChevronsUpDownIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search server..." />
          <CommandList>
            <CommandEmpty>No server found.</CommandEmpty>
            <CommandGroup>
              {guilds?.map((guild) => (
                <CommandItem
                  key={guild!.id}
                  value={guild!.id}
                  onSelect={(currentValue) => {
                    onChange(currentValue);
                    setOpen(false);
                  }}
                >
                  <CheckIcon
                    className={cn(
                      "mr-2 h-4 w-4",
                      value === guild!.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {guild!.name}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
