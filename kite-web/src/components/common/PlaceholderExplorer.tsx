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

const placeholderGroups = [
  {
    label: "User",
    placeholders: [
      "interaction.user.id",
      "interaction.user.mention",
      "interaction.user.username",
      "interaction.user.discriminator",
      "interaction.user.display_name",
      "interaction.user.avatar_url",
      "interaction.user.banner_url",
    ],
  },
  {
    label: "Command",
    placeholders: ["interaction.command.args.[name]"],
  },
  {
    label: "Nodes",
    placeholders: ["interaction.nodes.[id].result"],
  },
];

export default function PlaceholderExplorer({
  children,
  onSelect,
}: {
  children: ReactNode;
  onSelect: (value: string) => void;
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
            {placeholderGroups.map((group) => (
              <CommandGroup heading={group.label} key={group.label}>
                {group.placeholders.map((placeholder) => (
                  <CommandItem
                    key={placeholder}
                    value={placeholder}
                    onSelect={(currentValue) => {
                      onSelect(currentValue);
                      setOpen(false);
                    }}
                  >
                    <span className="text-muted-foreground mr-1">{"{{"}</span>
                    <span>{placeholder}</span>
                    <span className="text-muted-foreground ml-1">{"}}"}</span>
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
