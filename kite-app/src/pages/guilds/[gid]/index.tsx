import AppLayout from "@/components/AppLayout";
import { useGuildQuery } from "@/lib/api/queries";
import { guildIconUrl } from "@/lib/discord/cdn";
import { guildNameAbbreviation } from "@/lib/discord/util";
import { useRouteParams } from "@/hooks/route";

export default function GuildPage() {
  const { guildId } = useRouteParams();

  const { data: resp } = useGuildQuery(guildId);

  const guild = resp?.success ? resp.data : null;

  return (
    <AppLayout>
      <div className="mb-28 px-5 pb-16 border-b-4 border-slate-600 w-full">
        <div className="flex space-x-5 items-center">
          <div className="w-24 h-24 bg-slate-900 rounded-full flex items-center justify-center">
            {guild?.icon ? (
              <img
                src={guildIconUrl(guild)!}
                alt=""
                className="rounded-full h-full w-full"
              />
            ) : (
              <div className="text-2xl text-gray-300">
                {guildNameAbbreviation(guild?.name || "")}
              </div>
            )}
          </div>
          <div>
            <div className="text-xl font-medium text-gray-100 mb-2">
              {guild?.name || "Unknown Guild"}
            </div>
            <div className="text-gray-300 font-light">
              {guild?.description || "No description"}
            </div>
          </div>
        </div>
      </div>
    </AppLayout>
  );
}
