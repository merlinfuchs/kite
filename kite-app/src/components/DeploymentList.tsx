import { useDeploymentDeleteMutation } from "@/api/mutations";
import { useDeploymentsQuery, useWorkspacesQuery } from "@/api/queries";
import Link from "next/link";
import AutoAnimate from "./AutoAnimate";
import toast from "react-hot-toast";
import clsx from "clsx";

export default function DeploymentList({ guildId }: { guildId: string }) {
  const { data: resp, refetch } = useDeploymentsQuery(guildId);

  const deleteMutation = useDeploymentDeleteMutation();

  function deleteDeployment(deploymentId: string) {
    if (confirm("Are you sure you want to delete this workspace?")) {
      deleteMutation.mutate(
        { guildId, deploymentId },
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
              <div className="bg-slate-800 p-5 rounded-md" key={d.id}>
                <div className="flex mb-6">
                  <div className="flex-auto">
                    <div className="text-gray-100 text-lg font-medium mb-1">
                      {d.name}
                    </div>
                    <div className="font-light text-gray-300">
                      {d.description}
                    </div>
                  </div>
                  <div className="flex-none flex space-x-3 items-start">
                    <button
                      className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-gray-100 rounded"
                      onClick={() => deleteDeployment(d.id)}
                    >
                      Delete
                    </button>
                    <Link
                      className="px-3 py-2 bg-slate-700 hover:bg-slate-600 text-gray-100 rounded"
                      href={`/guilds/${guildId}/deployments/${d.id}`}
                    >
                      View Details
                    </Link>
                  </div>
                </div>
                <div className="bg-slate-900 rounded-md h-64"></div>
              </div>
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
