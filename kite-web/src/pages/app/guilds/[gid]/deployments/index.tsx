import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppDeploymentList from "@/components/app/AppDeploymentList";
import { useRouteParams } from "@/hooks/route";

export default function GuildDeploymentsPage() {
  const { guildId } = useRouteParams();

  return (
    <AppGuildLayout>
      <div>
        <div className="text-4xl font-bold text-white mb-4">Deployments</div>
        <div className="text-lg font-light text-gray-300 mb-10">
          A deployment is a running instance of a plugin. You can create a
          deployment from a workspace or by using a brebuilt plugin from the
          marketplace.
        </div>
        <AppDeploymentList guildId={guildId} />
      </div>
    </AppGuildLayout>
  );
}
