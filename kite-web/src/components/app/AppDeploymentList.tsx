import { useDeploymentDeleteMutation } from "@/lib/api/mutations";
import { useDeploymentsQuery } from "@/lib/api/queries";
import AutoAnimate from "../AutoAnimate";
import toast from "react-hot-toast";
import clsx from "clsx";
import AppDeploymentListEntry from "./AppDeploymentListEntry";
import AppIllustrationPlaceholder from "./AppIllustrationPlaceholder";
import { useState } from "react";
import ModalConfirm from "../ModalConfirm";
import AppDeploymentCreateModal from "./AppDeploymentCreateModal";

export default function AppDeploymentList({ appId }: { appId: string }) {
  const { data: resp } = useDeploymentsQuery(appId);

  const deployments = resp?.success ? resp.data : [];

  const [deleteDeploymentId, setDeleteDeploymentId] = useState<string | null>(
    null
  );

  const [createModalOpen, setCreateModalOpen] = useState(false);

  const deleteMutation = useDeploymentDeleteMutation(appId);

  function deleteDeployment() {
    if (!deleteDeploymentId) return;

    deleteMutation.mutate(
      { deploymentId: deleteDeploymentId },
      {
        onSuccess: (res) => {
          if (!res.success) {
            toast.error("Failed to delete workspace");
          } else {
            setDeleteDeploymentId(null);
          }
        },
      }
    );
  }

  return (
    <div>
      <AppDeploymentCreateModal
        open={createModalOpen}
        setOpen={setCreateModalOpen}
        appId={appId}
      />
      <ModalConfirm
        open={!!deleteDeploymentId}
        onCancel={() => setDeleteDeploymentId(null)}
        onConfirm={deleteDeployment}
        title="Delete Deployment"
        description="Are you sure you want to delete this deployment? This action is irreversible."
      />

      <div>
        <AutoAnimate
          className={clsx(
            "flex flex-col space-y-5",
            deployments.length !== 0 && "mb-10"
          )}
        >
          {deployments.map((d) => (
            <AppDeploymentListEntry
              key={d.id}
              appId={appId}
              deployment={d}
              onDelete={() => setDeleteDeploymentId(d.id)}
            />
          ))}
        </AutoAnimate>
        <div className="flex space-x-3">
          <button
            className="px-4 py-2 text-gray-100 rounded border-2 border-dark-9 hover:bg-dark-5 text-lg"
            onClick={() => setCreateModalOpen(true)}
          >
            New Deployment
          </button>
        </div>
      </div>
      {deployments.length === 0 && (
        <AppIllustrationPlaceholder
          svgPath="/illustrations/deploy.svg"
          title="Make your first deployment by creating a workspace or deploying a plugin from the marketplace!"
          className="mt-10"
        />
      )}
    </div>
  );
}
