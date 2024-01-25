import { Deployment } from "@/lib/api/wire";
import Link from "next/link";
import DeploymentLogSummary from "./DeploymentLogSummary";

interface Props {
  guildId: string;
  deployment: Deployment;
  onDelete: () => void;
}

import dynamic from "next/dynamic";

const DeploymentMetricsSummary = dynamic(
  () => import("./DeploymentMetricsSummary"),
  {
    ssr: false,
  }
);

export default function DeploymentListEntry({
  guildId,
  deployment,
  onDelete,
}: Props) {
  return (
    <div className="bg-slate-800 p-5 rounded-md">
      <div className="flex mb-6">
        <div className="flex-auto">
          <div className="text-gray-100 text-lg font-medium mb-1">
            {deployment.name}
          </div>
          <div className="font-light text-gray-300">
            {deployment.description}
          </div>
        </div>
        <div className="flex-none flex space-x-3 items-start">
          <button
            className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-gray-100 rounded"
            onClick={onDelete}
          >
            Delete
          </button>
          <Link
            className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-gray-100 rounded"
            href={`/guilds/${guildId}/deployments/${deployment.id}`}
          >
            View Details
          </Link>
        </div>
      </div>
      <div className="bg-slate-900 rounded-md flex flex-col px-1 py-2 space-y-1 mb-5">
        <DeploymentMetricsSummary
          guildId={guildId}
          deploymentId={deployment.id}
        />
      </div>
      <div>
        <DeploymentLogSummary guildId={guildId} deploymentId={deployment.id} />
      </div>
    </div>
  );
}