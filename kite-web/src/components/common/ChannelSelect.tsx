import { useAppStateGuildChannels } from "@/lib/hooks/api";
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
import { useMemo, useState } from "react";

// Text, voice, announcement, stage
const sendableChannelTypes = new Set([0, 2, 5, 13]);

export default function ChannelSelect({
  guildId,
  value,
  onChange,
}: {
  guildId: string | null;
  value: string | null;
  onChange: (value: string | null) => void;
}) {
  const allChannels = useAppStateGuildChannels(guildId);
  const channels = useMemo(
    () => allChannels?.filter((c) => c && sendableChannelTypes.has(c.type)),
    [allChannels]
  );

  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between truncate flex"
        >
          <div className="truncate">
            {value
              ? channels?.find((c) => c!.id === value)?.name
              : "Select channel..."}
          </div>
          <ChevronsUpDownIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search channel..." />
          <CommandList>
            <CommandEmpty>No channel found.</CommandEmpty>
            <CommandGroup>
              {channels?.map((channel) => (
                <CommandItem
                  key={channel!.id}
                  value={channel!.id}
                  keywords={[channel!.name]}
                  onSelect={(currentValue) => {
                    onChange(currentValue);
                    setOpen(false);
                  }}
                >
                  <CheckIcon
                    className={cn(
                      "mr-2 h-4 w-4",
                      value === channel!.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {channel!.name}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
