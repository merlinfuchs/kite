import AppLayout from "@/components/AppLayout";
import DeploymentLogs from "@/components/DeploymentLogs";
import { useRouteParams } from "@/hooks/route";

export default function DeploymentPage() {
  const { guildId, deploymentId } = useRouteParams();

  return (
    <AppLayout>
      <div className="p-5 bg-slate-800 rounded-md space-y-2 h-[500px] overflow-y-scroll flex flex-col justify-end">
        <DeploymentLogs guildId={guildId} deploymentId={deploymentId} />
      </div>
    </AppLayout>
  );
}