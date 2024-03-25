import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppDeploymentList from "@/components/app/AppDeploymentList";
import { useRouteParams } from "@/hooks/route";
import AppGuildPageHeader from "@/components/app/AppGuildPageHeader";

export default function GuildDeploymentsPage() {
  const { guildId } = useRouteParams();

  return (
    <AppGuildLayout>
      <AppGuildPageHeader
        title="Deployments"
        description="A deployment is a running instance of a plugin. You can create a
          deployment from a workspace or by using a brebuilt plugin from the
          marketplace."
      />
      <AppDeploymentList guildId={guildId} />
    </AppGuildLayout>
  );
}
