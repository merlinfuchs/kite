import AppLayout from "@/components/app/AppLayout";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
        <div className="text-lg text-gray-300 mb-10">
          To get started select the server from below that you want to use Kite
          on. If you haven't invited our bot yet, it will prompt you to do so.
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
          {guilds.map((guild) => (
            <Link key={guild.id} href={`/app/guilds/${guild.id}`}>
              <Card className="flex items-center hover:scale-101">
                <div className="bg-muted h-14 w-14 rounded-full flex items-center justify-center flex-none ml-5">
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
                <CardHeader>
                  <CardTitle>{guild.name}</CardTitle>
                  <CardDescription>{guild.id}</CardDescription>
                </CardHeader>
              </Card>
            </Link>
          ))}
          <a href={getApiUrl("/v1/auth/invite")}>
            <Card className="flex items-center h-full hover:scale-101 border-dashed">
              <PlusCircleIcon className="h-14 w-14 text-gray-400 group-hover:text-gray-300 ml-5" />
              <CardHeader>
                <CardTitle>Add to server</CardTitle>
                <CardDescription>Add Kite to a new server.</CardDescription>
              </CardHeader>
            </Card>
          </a>
        </div>
      </div>
    </AppLayout>
  );
}
