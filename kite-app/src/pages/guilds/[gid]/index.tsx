import AppLayout from "@/components/AppLayout";
import WorkspaceList from "@/components/WorkspaceList";
import DeploymentList from "@/components/DeploymentList";
import { useRouter } from "next/router";
import { useGuildQuery } from "@/api/queries";
import { guildIconUrl } from "@/discord/cdn";
import { guildNameAbbreviation } from "@/discord/util";

export default function GuildPage() {
  const router = useRouter();
  const guildId = router.query.gid as string;

  const { data: resp } = useGuildQuery(guildId);

  const guild = resp?.success ? resp.data : null;

  return (
    <AppLayout>
      <div className="mb-28 px-5 pb-10 border-b-4 border-slate-600">
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
      <div className="mb-28">
        <div className="text-4xl font-bold text-white mb-4">Deployments</div>
        <div className="text-lg font-light text-gray-300 mb-10">
          A deployment is a running instance of a plugin. You can create a
          deployment from a workspace or by using a brebuilt plugin from the
          marketplace.
        </div>
        <DeploymentList guildId={guildId} />
      </div>
      <div>
        <div className="text-4xl font-bold text-white mb-4">Workspaces</div>
        <div className="text-lg font-light text-gray-300 mb-10">
          A workspace is like a online VS Code project that can contain an
          arbitrary number of files and is used to create a private deployment
          or a public plugin.
        </div>
        <WorkspaceList guildId={guildId} />
      </div>
    </AppLayout>
  );
}
