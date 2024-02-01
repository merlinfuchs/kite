import { useGuildsQuery } from "@/lib/api/queries";
import { guildIconUrl } from "@/lib/discord/cdn";
import { guildNameAbbreviation } from "@/lib/discord/util";
import { ChevronDownIcon } from "@heroicons/react/20/solid";
import { Menu } from "@headlessui/react";
import clsx from "clsx";
import Link from "next/link";

export default function AppSidebarGuildSelect({
  guildId,
}: {
  guildId: string;
}) {
  const { data: resp } = useGuildsQuery();

  const guilds = resp?.success ? resp.data : [];
  const guild = guilds.find((g) => g.id === guildId);

  return (
    <Menu as="div" className="relative">
      <Menu.Button className="bg-dark-3 px-3 py-2 rounded cursor-pointer w-full hover:bg-dark-4">
        <div className="flex items-center select-none">
          <div className="bg-dark-1 h-10 w-10 rounded-full flex items-center justify-center flex-none mr-2">
            {guild?.icon ? (
              <img
                src={guildIconUrl(guild)!}
                alt={guildNameAbbreviation(guild?.name || "")}
                className="rounded-full h-full w-full"
              />
            ) : (
              <div className="text-base text-gray-300">
                {guildNameAbbreviation(guild?.name || "")}
              </div>
            )}
          </div>
          <div className="truncate text-gray-300 flex-auto text-left">
            {guild?.name || "Unknown Server"}
          </div>
          <div className="flex-none">
            <ChevronDownIcon
              className="h-5 w-5 text-gray-300"
              aria-hidden="true"
            />
          </div>
        </div>
      </Menu.Button>
      <Menu.Items className="absolute left-0 right-0 z-10 mt-2 w-full origin-top-right rounded bg-dark-3 shadow-lg focus:outline-none overflow-hidden">
        {guilds.map((g) => (
          <Menu.Item key={g.id}>
            {({ active }) => (
              <Link
                className={clsx(
                  active && "bg-dark-4",
                  "px-4 py-2 text-sm flex items-center space-x-2 cursor-pointer"
                )}
                href={`/app/guilds/${g.id}`}
              >
                <div className="bg-dark-1 h-10 w-10 rounded-full flex items-center justify-center flex-none mr-2">
                  {g?.icon ? (
                    <img
                      src={guildIconUrl(g)!}
                      alt={guildNameAbbreviation(g?.name || "")}
                      className="rounded-full h-full w-full"
                    />
                  ) : (
                    <div className="text-base text-gray-300">
                      {guildNameAbbreviation(g?.name || "")}
                    </div>
                  )}
                </div>
                <div className="truncate text-gray-300 flex-auto">
                  {g?.name || "Unknown Server"}
                </div>
              </Link>
            )}
          </Menu.Item>
        ))}
      </Menu.Items>
    </Menu>
  );
}
