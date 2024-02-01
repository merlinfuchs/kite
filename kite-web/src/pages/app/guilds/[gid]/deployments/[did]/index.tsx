import AppGuildLayout from "@/components/AppGuildLayout";
import DeploymentLogs from "@/components/DeploymentLogs";
import { useRouteParams } from "@/hooks/route";
import { useDeploymentQuery, useGuildQuery } from "@/lib/api/queries";
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

  const { data: resp } = useDeploymentQuery(guildId, deploymentId);

  const deployment = resp?.success ? resp.data : null;

  return (
    <AppGuildLayout>
      <div className="mb-10 bg-dark-2 p-5 rounded-md w-full">
        <div className="flex space-x-5 items-center mb-10">
          <div>
            <div className="text-xl font-medium text-gray-100 mb-2">
              {deployment?.name || "Unknown Deployment"}
            </div>
            <div className="text-gray-300 font-light">
              {deployment?.description || "No description"}
            </div>
          </div>
        </div>
        <div className="grid grid-cols-3">
          <div>
            <div className="text-gray-100 font-medium mb-1">Deployment ID</div>
            <div className="text-gray-300 text-sm">{deployment?.id}</div>
          </div>
          <div>
            <div className="text-gray-100 font-medium">Deployment Key</div>
            <div className="text-gray-300 text-sm">{deployment?.key}</div>
          </div>
        </div>
      </div>
      <div className="bg-dark-2 p-3 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-2">Logs</div>
        <div className="bg-dark-1 p-3 rounded-md h-64 overflow-y-auto space-y-2">
          <DeploymentLogs guildId={guildId} deploymentId={deploymentId} />
        </div>
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
    </AppGuildLayout>
  );
}
