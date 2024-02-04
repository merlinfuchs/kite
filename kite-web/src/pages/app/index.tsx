import AppLayout from "@/components/app/AppLayout";
import { getApiUrl } from "@/lib/api/client";
import { useGuildsQuery } from "@/lib/api/queries";
import { guildIconUrl } from "@/lib/discord/cdn";
import { guildNameAbbreviation } from "@/lib/discord/util";
import { PlusCircleIcon } from "@heroicons/react/24/outline";
import Link from "next/link";

export default function GuildsPage() {
  const { data: guildsResp } = useGuildsQuery();

  const guilds = guildsResp?.success ? guildsResp.data : [];

  return (
    <AppLayout>
      <div className="max-w-5xl mx-auto pb-20 pt-10 lg:pt-20 px-5">
        <div className="text-4xl font-bold text-white mb-4">
          Welcome to Kite!
        </div>
        <div className="text-lg font-light text-gray-300 mb-10">
          To get started select the server from below that you want to use Kite
          on. If you haven't invited our bot yet, it will prompt you to do so.
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
          {guilds.map((guild) => (
            <Link
              key={guild.id}
              className="bg-dark-2 rounded-md px-3 py-3 flex items-center hover:scale-101"
              href={`/app/guilds/${guild.id}`}
            >
              <div className="bg-dark-1 h-14 w-14 rounded-full flex items-center justify-center flex-none mr-4">
                {guild?.icon ? (
                  <img
                    src={guildIconUrl(guild)!}
                    alt={guildNameAbbreviation(guild?.name || "")}
                    className="rounded-full h-full w-full"
                  />
                ) : (
                  <div className="text-xl text-gray-300">
                    {guildNameAbbreviation(guild?.name || "")}
                  </div>
                )}
              </div>
              <div className="truncate">
                <div className="text-lg font-medium text-gray-100 mb-2 truncate">
                  {guild.name}
                </div>
                <div className="text-gray-400 text-sm truncate">{guild.id}</div>
              </div>
            </Link>
          ))}
          <a
            className="rounded-md px-3 py-3 border-2 border-dashed border-dark-7 hover:bg-dark-4 flex items-center"
            href={getApiUrl("/v1/auth/invite")}
          >
            <PlusCircleIcon className="h-14 w-14 text-gray-300 mr-3" />
            <div className="text-lg font-medium text-gray-100">
              Add to server
            </div>
          </a>
        </div>
      </div>
    </AppLayout>
  );
}
