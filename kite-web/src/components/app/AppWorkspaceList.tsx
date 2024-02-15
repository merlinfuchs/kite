import { useWorkspaceDeleteMutation } from "@/lib/api/mutations";
import { useWorkspacesQuery } from "@/lib/api/queries";
import Link from "next/link";
import AutoAnimate from "../AutoAnimate";
import toast from "react-hot-toast";
import clsx from "clsx";
import AppIllustrationPlaceholder from "./AppIllustrationPlaceholder";
import AppWorkspaceCreateModal from "./AppWorkspaceCreateModal";
import { useState } from "react";
import ModalConfirm from "@/components/ModalConfirm";

export default function AppWorkspaceList({ guildId }: { guildId: string }) {
  const { data: resp } = useWorkspacesQuery(guildId);

  const workspaces = resp?.success ? resp.data : [];

  const [deleteWorkspaceId, setDeleteWorkspaceId] = useState<string | null>(
    null
  );
  const [createModalOpen, setCreateModalOpen] = useState(false);

  const deleteMutation = useWorkspaceDeleteMutation(guildId);

  function deleteWorkspace() {
    if (!deleteWorkspaceId) return;

    deleteMutation.mutate(
      { workspaceId: deleteWorkspaceId },
      {
        onSuccess: (res) => {
          if (!res.success) {
            toast.error("Failed to delete workspace");
          } else {
            setDeleteWorkspaceId(null);
          }
        },
      }
    );
  }

  return (
    <div>
      <AppWorkspaceCreateModal
        open={createModalOpen}
        setOpen={setCreateModalOpen}
        guildId={guildId}
      />
      <ModalConfirm
        open={!!deleteWorkspaceId}
        onConfirm={deleteWorkspace}
        onCancel={() => setDeleteWorkspaceId(null)}
        title="Delete Workspace"
        description="Are you sure you want to delete this workspace? This action is irreversible."
      />

      <div>
        <AutoAnimate
          className={clsx(
            "flex flex-col space-y-5",
            workspaces.length !== 0 && "mb-10"
          )}
        >
          {workspaces.map((w) => (
            <div className="bg-dark-2 px-5 py-4 rounded-md" key={w.id}>
              <div className="flex">
                <div className="flex-auto">
                  <div className="text-gray-100 text-lg font-medium mb-1">
                    {w.name}
                  </div>
                  <div className="font-light text-gray-300">
                    {w.description}
                  </div>
                </div>
                <div className="flex-none flex space-x-3 items-start">
                  <button
                    className="px-3 py-2 bg-dark-4 hover:bg-dark-5 text-gray-100 rounded select-none"
                    onClick={() => setDeleteWorkspaceId(w.id)}
                  >
                    Delete
                  </button>
                  <Link
                    className="px-3 py-2 bg-dark-4 hover:bg-dark-5 text-gray-100 rounded select-none"
                    href={`/app/guilds/${guildId}/workspaces/${w.id}`}
                  >
                    Open Editor
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </AutoAnimate>
        <div className="flex space-x-3">
          <button
            className="px-4 py-2 text-gray-100 rounded border-2 border-dark-9 hover:bg-dark-5 text-lg"
            onClick={() => setCreateModalOpen(true)}
          >
            New Workspace
          </button>
        </div>
      </div>
      {workspaces.length === 0 && (
        <AppIllustrationPlaceholder
          svgPath="/illustrations/software_engineer.svg"
          title="Create you first workspace and get coding without worrying about the boring stuff!"
          className="mt-10"
        />
      )}
    </div>
  );
}
