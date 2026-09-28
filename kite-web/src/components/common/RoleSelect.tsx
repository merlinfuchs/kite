import { useAppStateGuildRoles } from "@/lib/hooks/api";
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

export default function RoleSelect({
  guildId,
  value,
  onChange,
}: {
  guildId: string | null;
  value: string | null;
  onChange: (value: string | null) => void;
}) {
  const allRoles = useAppStateGuildRoles(guildId);
  // Everyone and roles of integrations can't be given to members.
  const roles = useMemo(
    () =>
      allRoles
        ?.filter((r) => r && r.id !== guildId && !r.managed)
        .sort((a, b) => b!.position - a!.position),
    [allRoles, guildId]
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
              ? roles?.find((r) => r!.id === value)?.name
              : "Select role..."}
          </div>
          <ChevronsUpDownIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search role..." />
          <CommandList>
            <CommandEmpty>No role found.</CommandEmpty>
            <CommandGroup>
              {roles?.map((role) => (
                <CommandItem
                  key={role!.id}
                  value={role!.id}
                  keywords={[role!.name]}
                  onSelect={(currentValue) => {
                    onChange(currentValue);
                    setOpen(false);
                  }}
                >
                  <CheckIcon
                    className={cn(
                      "mr-2 h-4 w-4",
                      value === role!.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {role!.name}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
