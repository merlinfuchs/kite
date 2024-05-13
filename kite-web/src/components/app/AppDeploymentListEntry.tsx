import { Deployment } from "@/lib/types/wire";
import Link from "next/link";
import DeploymentLogSummary from "./AppDeploymentLogSummary";

interface Props {
  appId: string;
  deployment: Deployment;
  onDelete: () => void;
}

import dynamic from "next/dynamic";

const DeploymentMetricsEvents = dynamic(
  () => import("./AppDeploymentMetricsEvents"),
  {
    ssr: false,
  }
);

export default function AppDeploymentListEntry({
  appId,
  deployment,
  onDelete,
}: Props) {
  return (
    <div className="bg-dark-2 p-5 rounded-md">
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
            className="px-3 py-2 bg-dark-4 hover:bg-dark-5 text-gray-100 rounded"
            onClick={onDelete}
          >
            Delete
          </button>
          <Link
            className="px-3 py-2 bg-dark-4 hover:bg-dark-5 text-gray-100 rounded"
            href={`/apps/${appId}/deployments/${deployment.id}`}
          >
            View Details
          </Link>
        </div>
      </div>
      <div className="bg-dark-1 rounded-md flex flex-col px-1 py-2 space-y-1 mb-5">
        <DeploymentMetricsEvents appId={appId} deploymentId={deployment.id} />
      </div>
      <div>
        <DeploymentLogSummary appId={appId} deploymentId={deployment.id} />
      </div>
    </div>
  );
}
