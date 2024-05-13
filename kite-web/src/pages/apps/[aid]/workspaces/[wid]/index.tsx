import {
  useCompileMutation,
  useDeploymentCreateMutation,
  useWorkspaceUpdateMutation,
} from "@/lib/api/mutations";
import { useWorkspaceQuery } from "@/lib/api/queries";
import { Workspace } from "@/lib/types/wire";
import Code from "@/components/code/Code";
import {
  compileWorkspace,
  readManifestFromWorkspace,
} from "@/lib/code/compile";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import { useRouteParams } from "@/hooks/route";
import Flow from "@/components/flow/Flow";
import { bundleFlowFiles } from "@/lib/flow/bundle";

export default function AppWorkspacePage() {
  const router = useRouter();
  const { appId, workspaceId } = useRouteParams();

  const workspaceQuery = useWorkspaceQuery(appId, workspaceId);
  const updateWorkspaceMutation = useWorkspaceUpdateMutation(appId);

  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [openFilePath, setOpenFilePath] = useState<string | null>("index.ts");

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (workspaceQuery.data && workspaceQuery.data.success) {
      setWorkspace(workspaceQuery.data.data);
      setOpenFilePath("index.ts");

      setTimeout(() => {
        // This is a workaround to ignore the initial onChange triggered by reactflow
        setHasUnsavedChanges(false);
      }, 100);
    }
  }, [workspaceQuery.data]);

  async function onSave() {
    if (!workspace || !hasUnsavedChanges) return;

    setIsSaving(true);
    updateWorkspaceMutation.mutate(
      {
        workspaceId,
        req: {
          name: workspace.name,
          description: workspace.description,
          files: workspace.files.map((file) => ({
            path: file.path,
            content: file.content,
          })),
        },
      },
      {
        onSuccess: (res) => {
          setIsSaving(false);
          if (res.success) {
            setHasUnsavedChanges(false);
          }
        },
        onError: () => {
          setIsSaving(false);
        },
      }
    );
  }

  const [isDeploying, setIsDeploying] = useState(false);
  const compileMutation = useCompileMutation();
  const deployMutation = useDeploymentCreateMutation(appId);

  async function onDeploy() {
    if (!workspace) return;

    setIsDeploying(true);

    let bundledSource: string;

    if (workspace.type === "FLOW") {
      const bundledFlow = bundleFlowFiles(workspace.files);
      bundledSource = JSON.stringify(bundledFlow);
    } else if (workspace.type === "JS") {
      const bundledJs = await compileWorkspace(workspace.files, "index.ts");
      bundledSource = bundledJs;
    } else {
      console.warn("Unknown workspace type");
      return;
    }

    const manifest = readManifestFromWorkspace(workspace.files);

    compileMutation.mutate(
      {
        type: workspace.type,
        source: bundledSource,
      },
      {
        onSuccess: (res) => {
          if (!res.success) {
            console.log(res.error);
            setIsDeploying(false);
            return;
          }

          deployMutation.mutate(
            {
              key: manifest?.deployment?.key || "default@web",
              name: manifest?.deployment?.name || "Untitled Plugin",
              description:
                manifest?.deployment?.description || "No description",
              wasm_bytes: res.data.wasm_bytes,
              plugin_version_id: null,
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

  if (workspace.type === "FLOW") {
    return (
      <Flow
        files={workspace.files}
        openFilePath="default.flow"
        setOpenFilePath={() => {}}
        isSaving={isSaving}
        isDeploying={isDeploying}
        hasUnsavedChanges={hasUnsavedChanges}
        onChange={() => setHasUnsavedChanges(true)}
        onDeploy={onDeploy}
        onSave={onSave}
        onExit={() => router.push(`/apps/${appId}/workspaces`)}
      />
    );
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
        onExit={() => router.push(`/apps/${appId}/workspaces`)}
        isDeploying={isDeploying}
        onDeploy={onDeploy}
      />
    </div>
  );
}
