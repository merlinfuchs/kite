import { useWorkspaceDeleteMutation } from "@/lib/api/mutations";
import { useWorkspacesQuery } from "@/lib/api/queries";
import Link from "next/link";
import AutoAnimate from "../AutoAnimate";
import { toast } from "sonner";
import clsx from "clsx";
import AppWorkspaceCreateDialog from "./AppWorkspaceCreateDialog";
import { useState } from "react";
import ConfirmDialog from "@/components/ConfirmDialog";
import { Card, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { Button } from "../ui/button";
import AppGuildPageEmpty from "./AppGuildPageEmpty";

export default function AppWorkspaceList({ guildId }: { guildId: string }) {
  const { data: resp } = useWorkspacesQuery(guildId);

  const workspaces = resp?.success ? resp.data : [];

  const [deleteWorkspaceId, setDeleteWorkspaceId] = useState<string | null>(
    null
  );

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
    <div className="flex flex-col flex-1">
      <ConfirmDialog
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
            <Card key={w.id}>
              <div className="flex flex-col md:flex-row justify-between">
                <CardHeader>
                  <CardTitle>{w.name}</CardTitle>
                  <CardDescription>{w.description}</CardDescription>
                </CardHeader>
                <div className="flex-none flex space-x-3 px-6 pb-6 md:pt-6">
                  <Button
                    variant="outline"
                    onClick={() => setDeleteWorkspaceId(w.id)}
                  >
                    Delete
                  </Button>
                  <Button asChild variant="secondary">
                    <Link href={`/app/guilds/${guildId}/workspaces/${w.id}`}>
                      Open Editor
                    </Link>
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </AutoAnimate>
      </div>

      {workspaces.length === 0 ? (
        <AppGuildPageEmpty
          title="There are no workspaces"
          description="Create a new workspace to start building your plugin."
          action={<AppWorkspaceCreateDialog guildId={guildId} />}
        />
      ) : (
        <div>
          <AppWorkspaceCreateDialog guildId={guildId} />
        </div>
      )}
    </div>
  );
}
