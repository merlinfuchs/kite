import AppLayout from "@/components/AppLayout";
import DeploymentLogs from "@/components/DeploymentLogs";
import { useRouteParams } from "@/hooks/route";
import dynamic from "next/dynamic";

const DeploymentMetricsEvents = dynamic(
  () => import("@/components/DeploymentMetricsEvents"),
  {
    ssr: false,
  }
);

const DeploymentMetricsCalls = dynamic(
  () => import("@/components/DeploymentMetricsCalls"),
  {
    ssr: false,
  }
);

const DeploymentMetricsTotalTime = dynamic(
  () => import("@/components/DeploymentMetricsTotalTime"),
  {
    ssr: false,
  }
);

const DeploymentMetricsExecutionTime = dynamic(
  () => import("@/components/DeploymentMetricsExecutionTime"),
  {
    ssr: false,
  }
);

export default function GuildDeploymentPage() {
  const { guildId, deploymentId } = useRouteParams();

  return (
    <AppLayout>
      <div className="p-5 bg-dark-2 rounded-md space-y-2 h-[500px] overflow-y-scroll mb-5">
        <DeploymentLogs guildId={guildId} deploymentId={deploymentId} />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Events Handled
        </div>
        <DeploymentMetricsEvents
          guildId={guildId}
          deploymentId={deploymentId}
        />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Actions Taken
        </div>
        <DeploymentMetricsCalls guildId={guildId} deploymentId={deploymentId} />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Average Total Time
        </div>
        <DeploymentMetricsTotalTime
          guildId={guildId}
          deploymentId={deploymentId}
        />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Average CPU Time
        </div>
        <DeploymentMetricsExecutionTime
          guildId={guildId}
          deploymentId={deploymentId}
        />
      </div>
    </AppLayout>
  );
}
