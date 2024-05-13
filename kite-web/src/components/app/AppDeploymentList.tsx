import { useDeploymentDeleteMutation } from "@/lib/api/mutations";
import { useDeploymentsQuery } from "@/lib/api/queries";
import AutoAnimate from "../AutoAnimate";
import toast from "react-hot-toast";
import clsx from "clsx";
import AppDeploymentListEntry from "./AppDeploymentListEntry";
import AppIllustrationPlaceholder from "./AppIllustrationPlaceholder";
import { useState } from "react";
import ModalConfirm from "../ModalConfirm";

export default function AppDeploymentList({ appId }: { appId: string }) {
  const { data: resp } = useDeploymentsQuery(appId);

  const deployments = resp?.success ? resp.data : [];

  const [deleteDeploymentId, setDeleteDeploymentId] = useState<string | null>(
    null
  );

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
