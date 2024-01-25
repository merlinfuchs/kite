import {
  useCompileJsMutation,
  useDeploymentCreateMutation,
  useWorkspaceUpdateMutation,
} from "@/lib/api/mutations";
import { useWorkspaceQuery } from "@/lib/api/queries";
import { Workspace } from "@/lib/api/wire";
import Code from "@/components/code/Code";
import { compileWorkspace, readManifestFromWorkspace } from "@/util/compile";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import { useRouteParams } from "@/hooks/route";

export default function GuildWorkspacePage() {
  const router = useRouter();
  const { guildId, workspaceId } = useRouteParams();

  const workspaceQuery = useWorkspaceQuery(guildId, workspaceId);
  const updateWorkspaceMutation = useWorkspaceUpdateMutation(guildId);

  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [openFilePath, setOpenFilePath] = useState<string | null>("index.ts");

  useEffect(() => {
    if (workspaceQuery.data && workspaceQuery.data.success) {
      setWorkspace(workspaceQuery.data.data);
      setOpenFilePath("index.ts");
    }
  }, [workspaceQuery.data]);

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  async function onSave() {
    if (!workspace || !hasUnsavedChanges) return;

    setIsSaving(true);
    updateWorkspaceMutation.mutate(
      {
        workspaceId,
        req: {
          name: "Some Workspace",
          description: "Some description",
          files: workspace.files.map((file) => ({
            path: file.path,
            content: file.content,
          })),
        },
      },
      {
        onSuccess: () => {
          setIsSaving(false);
          setHasUnsavedChanges(false);
        },
        onError: () => {
          setIsSaving(false);
        },
      }
    );
  }

  const [isDeploying, setIsDeploying] = useState(false);
  const compileMutation = useCompileJsMutation();
  const deployMutation = useDeploymentCreateMutation(guildId);

  async function onDeploy() {
    if (!workspace) return;

    const bundledJs = await compileWorkspace(workspace.files, "index.ts");
    const manifest = readManifestFromWorkspace(workspace.files);

    setIsDeploying(true);

    compileMutation.mutate(
      {
        source: bundledJs,
      },
      {
        onSuccess: (res) => {
          if (!res.success) {
            console.log(res.error);
            return;
          }

          deployMutation.mutate(
            {
              key: manifest?.plugin?.key || "default@web",
              name: manifest?.plugin?.name || "Untitled Plugin",
              description: manifest?.plugin?.description || "No description",
              wasm_bytes: res.data.wasm_bytes,
              plugin_version_id: null,
              manifest_events: manifest?.plugin?.events || [],
              manifest_commands: [],
              manifest_default_config: {},
              config: {},
            },
            {
              onSettled: () => {
                setIsDeploying(false);
              },
            }
          );
        },
        onError: () => {
          setIsDeploying(false);
        },
      }
    );
  }

  if (!workspace) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <Code
        files={workspace.files}
        openFilePath={openFilePath}
        setOpenFilePath={setOpenFilePath}
        hasUnsavedChanges={hasUnsavedChanges}
        isSaving={isSaving}
        onSave={onSave}
        onChange={() => setHasUnsavedChanges(true)}
        onBack={() => router.push(`/guilds/${guildId}/workspaces`)}
        isDeploying={isDeploying}
        onDeploy={onDeploy}
      />
    </div>
  );
}
