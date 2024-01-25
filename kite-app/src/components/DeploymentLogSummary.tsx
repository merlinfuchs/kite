import { useDeploymentLogSummaryQuery } from "@/lib/api/queries";

interface Props {
  guildId: string;
  deploymentId: string;
}

export default function DeploymentLogSummary({ guildId, deploymentId }: Props) {
  const { data: resp } = useDeploymentLogSummaryQuery(guildId, deploymentId);

  const summary = resp?.success ? resp.data : null;

  return (
    <div className="flex text-gray-100 flex-wrap">
      <div className="text-gray-300 mr-4 mb-2">Last 24 hours</div>
      <div className="flex space-x-2 font-light mr-4 mb-2">
        <div className="bg-red-500 px-1 rounded">ERROR</div>
        <div>{summary?.error_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light mr-4 mb-2">
        <div className="bg-yellow-500 px-1 rounded">WARN</div>
        <div>{summary?.warn_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light mr-4 mb-2">
        <div className="bg-blue-500 px-1 rounded">INFO</div>
        <div>{summary?.info_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light mr-4 mb-2">
        <div className="bg-gray-700 px-1 rounded">DEBUG</div>
        <div>{summary?.debug_count || 0}</div>
      </div>
    </div>
  );
}
