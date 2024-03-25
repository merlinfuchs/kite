import { useGuildsQuery } from "@/lib/api/queries";
import { guildIconUrl } from "@/lib/discord/cdn";
import { guildNameAbbreviation } from "@/lib/discord/util";
import { ChevronDownIcon } from "@heroicons/react/20/solid";
import Link from "next/link";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";

export default function AppGuildSelect({ guildId }: { guildId: string }) {
  const { data: resp } = useGuildsQuery();

  const guilds = resp?.success ? resp.data : [];
  const guild = guilds.find((g) => g.id === guildId);

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="w-full">
        <div className="flex items-center select-none rounded-xl border bg-card text-card-foreground shadow p-2">
          <div className="bg-background h-10 w-10 rounded-full flex items-center justify-center flex-none mr-2">
            {guild?.icon ? (
              <img
                src={guildIconUrl(guild)!}
                alt={guildNameAbbreviation(guild?.name || "")}
                className="rounded-full h-full w-full"
              />
            ) : (
              <div className="text-base text-foreground">
                {guildNameAbbreviation(guild?.name || "")}
              </div>
            )}
          </div>
          <div className="truncate text-foreground flex-auto text-left mr-2">
            {guild?.name || "Unknown Server"}
          </div>
          <div className="flex-none">
            <ChevronDownIcon
              className="h-5 w-5 text-foreground"
              aria-hidden="true"
            />
          </div>
        </div>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        {guilds.map((g) => (
          <DropdownMenuItem key={g.id} asChild>
            <Link
              className="text-sm flex items-center space-x-2 cursor-pointer w-full"
              href={`/app/guilds/${g.id}`}
            >
              <div className="bg-muted h-10 w-10 rounded-full flex items-center justify-center flex-none mr-2">
                {g?.icon ? (
                  <img
                    src={guildIconUrl(g)!}
                    alt={guildNameAbbreviation(g?.name || "")}
                    className="rounded-full h-full w-full"
                  />
                ) : (
                  <div className="text-base text-foreground">
                    {guildNameAbbreviation(g?.name || "")}
                  </div>
                )}
              </div>
              <div className="truncate text-foreground flex-auto">
                {g?.name || "Unknown Server"}
              </div>
            </Link>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
