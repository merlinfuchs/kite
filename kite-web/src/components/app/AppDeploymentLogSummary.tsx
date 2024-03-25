import { useDeploymentLogSummaryQuery } from "@/lib/api/queries";
import { Badge } from "../ui/badge";

interface Props {
  guildId: string;
  deploymentId: string;
}

export default function DeploymentLogSummary({ guildId, deploymentId }: Props) {
  const { data: resp } = useDeploymentLogSummaryQuery(guildId, deploymentId);

  const summary = resp?.success ? resp.data : null;

  return (
    <div className="flex text-gray-100 flex-wrap gap-x-4 gap-y-3">
      <div className="text-gray-300">Last 24 hours</div>
      <div className="flex space-x-2 font-light">
        <Badge className="bg-red-500 uppercase">Error</Badge>
        <div>{summary?.error_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light">
        <Badge className="bg-yellow-500 uppercase">Warn</Badge>
        <div>{summary?.warn_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light">
        <Badge className="bg-blue-500 uppercase">Info</Badge>
        <div>{summary?.info_count || 0}</div>
      </div>
      <div className="flex space-x-2 font-light">
        <Badge className="bg-gray-700 uppercase">Debug</Badge>
        <div>{summary?.debug_count || 0}</div>
      </div>
    </div>
  );
}
