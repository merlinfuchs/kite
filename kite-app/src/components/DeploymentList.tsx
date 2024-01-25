import { useDeploymentDeleteMutation } from "@/lib/api/mutations";
import { useDeploymentsQuery, useWorkspacesQuery } from "@/lib/api/queries";
import Link from "next/link";
import AutoAnimate from "./AutoAnimate";
import toast from "react-hot-toast";
import clsx from "clsx";
import DeploymentListEntry from "./DeploymentListEntry";

export default function DeploymentList({ guildId }: { guildId: string }) {
  const { data: resp, refetch } = useDeploymentsQuery(guildId);

  const deleteMutation = useDeploymentDeleteMutation(guildId);

  function deleteDeployment(deploymentId: string) {
    if (confirm("Are you sure you want to delete this workspace?")) {
      deleteMutation.mutate(
        { deploymentId },
        {
          onSuccess: (res) => {
            if (res.success) {
              refetch();
            } else {
              toast.error("Failed to delete workspace");
            }
          },
        }
      );
    }
  }

  return (
    <div>
      {resp?.success ? (
        <div>
          <AutoAnimate
            className={clsx(
              "flex flex-col space-y-5",
              resp.data.length !== 0 && "mb-10"
            )}
          >
            {resp.data.map((d) => (
              <DeploymentListEntry
                key={d.id}
                guildId={guildId}
                deployment={d}
                onDelete={() => deleteDeployment(d.id)}
              />
            ))}
          </AutoAnimate>
          <div>
            <button
              className="px-4 py-2 text-gray-100 rounded border-2 border-slate-400 hover:bg-slate-600 text-lg"
              onClick={() => {}}
            >
              New Deployment
            </button>
          </div>
        </div>
      ) : (
        <div>Loading ...</div>
      )}
    </div>
  );
}
