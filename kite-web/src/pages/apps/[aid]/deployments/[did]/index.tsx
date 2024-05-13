import AppLayout from "@/components/app/AppLayout";
import AppDeploymentLogs from "@/components/app/AppDeploymentLogs";
import { useRouteParams } from "@/hooks/route";
import { useDeploymentQuery } from "@/lib/api/queries";
import dynamic from "next/dynamic";

const AppDeploymentMetricsEvents = dynamic(
  () => import("@/components/app/AppDeploymentMetricsEvents"),
  {
    ssr: false,
  }
);

const AppDeploymentMetricsCalls = dynamic(
  () => import("@/components/app/AppDeploymentMetricsCalls"),
  {
    ssr: false,
  }
);

const AppDeploymentMetricsTotalTime = dynamic(
  () => import("@/components/app/AppDeploymentMetricsTotalTime"),
  {
    ssr: false,
  }
);

const AppDeploymentMetricsExecutionTime = dynamic(
  () => import("@/components/app/AppDeploymentMetricsExecutionTime"),
  {
    ssr: false,
  }
);

export default function AppDeploymentPage() {
  const { appId, deploymentId } = useRouteParams();

  const { data: resp } = useDeploymentQuery(appId, deploymentId);

  const deployment = resp?.success ? resp.data : null;

  return (
    <AppLayout>
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
          <AppDeploymentLogs appId={appId} deploymentId={deploymentId} />
        </div>
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Events Handled
        </div>
        <AppDeploymentMetricsEvents appId={appId} deploymentId={deploymentId} />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Actions Taken
        </div>
        <AppDeploymentMetricsCalls appId={appId} deploymentId={deploymentId} />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Average Total Time
        </div>
        <AppDeploymentMetricsTotalTime
          appId={appId}
          deploymentId={deploymentId}
        />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Average CPU Time
        </div>
        <AppDeploymentMetricsExecutionTime
          appId={appId}
          deploymentId={deploymentId}
        />
      </div>
    </AppLayout>
  );
}
