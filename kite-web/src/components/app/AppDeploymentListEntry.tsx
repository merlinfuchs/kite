import { Deployment } from "@/lib/types/wire";
import Link from "next/link";
import DeploymentLogSummary from "./AppDeploymentLogSummary";

interface Props {
  guildId: string;
  deployment: Deployment;
  onDelete: () => void;
}

import dynamic from "next/dynamic";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Button } from "../ui/button";

const DeploymentMetricsEvents = dynamic(
  () => import("./AppDeploymentMetricsEvents"),
  {
    ssr: false,
  }
);

export default function AppDeploymentListEntry({
  guildId,
  deployment,
  onDelete,
}: Props) {
  return (
    <Card>
      <div className="flex flex-col md:flex-row justify-between">
        <CardHeader className="flex justify-between">
          <CardTitle>{deployment.name}</CardTitle>
          <CardDescription>{deployment.description}</CardDescription>
        </CardHeader>
        <div className="flex-none flex space-x-3 px-6 md:py-6">
          <Button variant="outline" onClick={onDelete}>
            Delete
          </Button>
          <Button asChild variant="secondary">
            <Link href={`/app/guilds/${guildId}/deployments/${deployment.id}`}>
              View Details
            </Link>
          </Button>
        </div>
      </div>
      <div className="flex mb-6"></div>
      <CardContent>
        <div className="mb-6">
          <DeploymentMetricsEvents
            guildId={guildId}
            deploymentId={deployment.id}
          />
        </div>
        <div>
          <DeploymentLogSummary
            guildId={guildId}
            deploymentId={deployment.id}
          />
        </div>
      </CardContent>
    </Card>
  );
}
