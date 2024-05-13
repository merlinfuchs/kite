import AppLayout from "@/components/app/AppLayout";
import { userAvatarUrl } from "@/lib/discord/cdn";
import { useRouteParams } from "@/hooks/route";
import dynamic from "next/dynamic";
import { useAppQuery } from "@/lib/api/queries";
import { nameAbbreviation } from "@/lib/discord/util";

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

const AppUsageSummary = dynamic(
  () => import("@/components/app/AppUsageSummary"),
  {
    ssr: false,
  }
);

export default function AppPage() {
  const { appId } = useRouteParams();

  const { data: resp } = useAppQuery(appId);

  const app = resp?.success ? resp.data : null;

  return (
    <AppLayout>
      <div className="mb-10 bg-dark-2 p-5 rounded-md w-full flex">
        <div className="flex-auto">
          <div className="flex space-x-5 items-center mb-10">
            <div className="w-24 h-24 bg-dark-1 rounded-full flex items-center justify-center">
              {app && (
                <img
                  src={
                    userAvatarUrl({
                      id: app.user_id,
                      discriminator: app.user_discriminator,
                      avatar: app.user_avatar,
                    })!
                  }
                  alt={nameAbbreviation(app.user_name)}
                  className="rounded-full h-full w-full"
                />
              )}
            </div>
            <div>
              <div className="text-xl font-medium text-gray-100 mb-2">
                {app?.user_name || "Unknown App"}
              </div>
              <div className="text-gray-300 font-light">
                {app?.user_bio || "No description"}
              </div>
            </div>
          </div>
          <div className="grid grid-cols-3">
            <div>
              <div className="text-gray-100 font-medium mb-1">App ID</div>
              <div className="text-gray-300 text-sm">{app?.id}</div>
            </div>
            <div>
              <div className="text-gray-100 font-medium">Servers</div>
              <div className="text-gray-300 text-sm">{0}</div>
            </div>
          </div>
        </div>
        <div>
          <AppUsageSummary appId={appId} />
        </div>
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Events Handled
        </div>
        <AppDeploymentMetricsEvents appId={appId} />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Actions Taken
        </div>
        <AppDeploymentMetricsCalls appId={appId} />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md mb-5">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Average Total Time
        </div>
        <AppDeploymentMetricsTotalTime appId={appId} />
      </div>
      <div className="bg-dark-2 px-1 py-2 rounded-md">
        <div className="text-gray-100 font-bold text-2xl mb-5 mx-5 mt-3">
          Average CPU Time
        </div>
        <AppDeploymentMetricsExecutionTime appId={appId} />
      </div>
    </AppLayout>
  );
}
