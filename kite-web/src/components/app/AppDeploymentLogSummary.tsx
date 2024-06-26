import { useDeploymentLogSummaryQuery } from "@/lib/api/queries";

interface Props {
  appId: string;
  deploymentId: string;
}

export default function DeploymentLogSummary({ appId, deploymentId }: Props) {
  const { data: resp } = useDeploymentLogSummaryQuery(appId, deploymentId);

  const summary = resp?.success ? resp.data : null;

  return (
    <div className="flex text-gray-100 flex-wrap gap-x-4 gap-y-3">
      <div className="text-gray-300">Last 24 hours</div>
      <div className="flex space-x-2 font-light">
        <div className="bg-red-500 px-1 rounded">ERROR</div>
        <div>{summary?.error_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light">
        <div className="bg-yellow-500 px-1 rounded">WARN</div>
        <div>{summary?.warn_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light">
        <div className="bg-blue-500 px-1 rounded">INFO</div>
        <div>{summary?.info_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light">
        <div className="bg-gray-700 px-1 rounded">DEBUG</div>
        <div>{summary?.debug_count || 0}</div>
      </div>
    </div>
  );
}
