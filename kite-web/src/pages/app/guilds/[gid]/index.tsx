import AppGuildLayout from "@/components/app/AppGuildLayout";
import { useGuildQuery } from "@/lib/api/queries";
import { guildIconUrl } from "@/lib/discord/cdn";
import { guildNameAbbreviation } from "@/lib/discord/util";
import { useRouteParams } from "@/hooks/route";
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

const AppGuildUsageSummary = dynamic(
  () => import("@/components/app/AppGuildUsageSummary"),
  {
    ssr: false,
  }
);

export default function GuildPage() {
  const { guildId } = useRouteParams();

  const { data: resp } = useGuildQuery(guildId);

  const guild = resp?.success ? resp.data : null;

  return (
    <AppGuildLayout>
      <div className="mb-10 bg-dark-2 p-5 rounded-md w-full flex">
        <div className="flex-auto">
          <div className="flex space-x-5 items-center mb-10">
            <div className="w-24 h-24 bg-dark-1 rounded-full flex items-center justify-center">
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
          <div className="grid grid-cols-3">
            <div>
              <div className="text-gray-100 font-medium mb-1">Guild ID</div>
              <div className="text-gray-300 text-sm">{guild?.id}</div>
            </div>
            <div>
              <div className="text-gray-100 font-medium">Members</div>
              <div className="text-gray-300 text-sm">{0}</div>
            </div>
          </div>
        </div>
        <div>
          <AppGuildUsageSummary guildId={guildId} />
        </div>
      </div>
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Events Handled</CardTitle>
            <CardDescription>
              Events that have been handled by deployments on this server.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsEvents guildId={guildId} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Actions Taken</CardTitle>
            <CardDescription>
              Actions takens by deployments on this server.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsCalls guildId={guildId} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Average Total Time</CardTitle>
            <CardDescription>
              Average time spent by deployments on this server to process events
              including waiting for actions to finish.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsTotalTime guildId={guildId} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Average Execution Time</CardTitle>
            <CardDescription>
              Average time spent by deployments on this server to execute code
              excluding waiting for actions to finish.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AppDeploymentMetricsExecutionTime guildId={guildId} />
          </CardContent>
        </Card>
      </div>
    </AppGuildLayout>
  );
}
