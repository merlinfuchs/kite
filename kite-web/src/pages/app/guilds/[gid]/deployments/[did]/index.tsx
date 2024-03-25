import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppDeploymentLogs from "@/components/app/AppDeploymentLogs";
import { useRouteParams } from "@/hooks/route";
import { useDeploymentQuery } from "@/lib/api/queries";
import dynamic from "next/dynamic";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

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
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Logs</CardTitle>
            <CardDescription>
              Log messages and errors for this deployment.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="p-3 h-64 overflow-y-auto space-y-2">
              <AppDeploymentLogs
                guildId={guildId}
                deploymentId={deploymentId}
              />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Events Handled</CardTitle>
            <CardDescription>
              Events that have been handled by this deployment.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsEvents
              guildId={guildId}
              deploymentId={deploymentId}
            />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Actions Taken</CardTitle>
            <CardDescription>
              Actions that have been taken by this deployment.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsCalls
              guildId={guildId}
              deploymentId={deploymentId}
            />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Actions Taken</CardTitle>
            <CardDescription>
              Average time spent by this deployment to process events including
              waiting for actions to finish.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsTotalTime
              guildId={guildId}
              deploymentId={deploymentId}
            />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Average Execution Time</CardTitle>
            <CardDescription>
              Average time spent by this deployment to execute code excluding
              waiting for actions to finish.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsExecutionTime
              guildId={guildId}
              deploymentId={deploymentId}
            />
          </CardContent>
        </Card>
      </div>
    </AppGuildLayout>
  );
}
