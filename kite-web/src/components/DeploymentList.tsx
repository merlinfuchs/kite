import { useDeploymentDeleteMutation } from "@/lib/api/mutations";
import { useDeploymentsQuery, useWorkspacesQuery } from "@/lib/api/queries";
import Link from "next/link";
import AutoAnimate from "./AutoAnimate";
import toast from "react-hot-toast";
import clsx from "clsx";
import DeploymentListEntry from "./DeploymentListEntry";
import IllustrationPlaceholder from "./IllustrationPlaceholder";

export default function DeploymentList({ guildId }: { guildId: string }) {
  const { data: resp } = useDeploymentsQuery(guildId);

  const deployments = resp?.success ? resp.data : [];

  const deleteMutation = useDeploymentDeleteMutation(guildId);

  function deleteDeployment(deploymentId: string) {
    if (confirm("Are you sure you want to delete this workspace?")) {
      deleteMutation.mutate(
        { deploymentId },
        {
          onSuccess: (res) => {
            if (!res.success) {
              toast.error("Failed to delete workspace");
            }
          },
        }
      );
    }
  }

  return (
    <div>
      <div>
        <AutoAnimate
          className={clsx(
            "flex flex-col space-y-5",
            deployments.length !== 0 && "mb-10"
          )}
        >
          {deployments.map((d) => (
            <DeploymentListEntry
              key={d.id}
              guildId={guildId}
              deployment={d}
              onDelete={() => deleteDeployment(d.id)}
            />
          ))}
        </AutoAnimate>
      </div>
      {deployments.length === 0 && (
        <IllustrationPlaceholder
          svgPath="/illustrations/deploy.svg"
          title="Make your first deployment by creating a workspace or deploying a plugin from the marketplace!"
          className="mt-10"
        />
      )}
    </div>
  );
}
