import { CommandDeployDialog } from "@/components/app/CommandDeployDialog";
import FlowPage from "@/components/flow/FlowPage";
import {
  useCommandsDeployMutation,
  useCommandUpdateMutation,
} from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/dataSchema";
import { useCommand, useFlowLogEntries } from "@/lib/hooks/api";
import { useUnsavedChangesWarning } from "@/lib/hooks/exit";
import { useAppId, useCommandId } from "@/lib/hooks/params";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";

export default function AppCommandPage() {
  const router = useRouter();
  const cmd = useCommand((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load command: ${res?.error.message} (${res?.error.code})`
      );
      if (res.error.code === "unknown_command") {
        router.push({
          pathname: "/apps/[appId]/commands",
          query: { appId: router.query.appId },
        });
      }
    }
  });

  const updateMutation = useCommandUpdateMutation(useAppId(), useCommandId());

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [deployDialogOpen, setDeployDialogOpen] = useState(false);

  const onChange = useCallback(() => {
    setHasUnsavedChanges(true);
  }, [setHasUnsavedChanges]);

  const save = useCallback(
    (data: FlowData, options?: { onSuccess?: () => void }) => {
      updateMutation.mutate(
        {
          flow_source: data,
        },
        {
          onSuccess(res) {
            if (res.success) {
              toast.success(
                "Command saved! Make sure to deploy for changes to take effect in Discord.",
                {
                  action: {
                    label: "Deploy",
                    onClick: () => setDeployDialogOpen(true),
                  },
                }
              );
              options?.onSuccess?.();
            } else {
              toast.error(
                `Failed to update command: ${res.error.message} (${res.error.code})`
              );
            }
          },
          onSettled() {
            setHasUnsavedChanges(false);
          },
        }
      );
    },
    [setHasUnsavedChanges, updateMutation]
  );

  const hasUndeployedChanges = useMemo(() => {
    return (
      cmd && new Date(cmd!.updated_at) > new Date(cmd!.last_deployed_at || 0)
    );
  }, [cmd]);

  const leave = useUnsavedChangesWarning(hasUnsavedChanges);

  const exit = useCallback(() => {
    leave({
      pathname: "/apps/[appId]/commands",
      query: { appId: router.query.appId },
    });
  }, [leave, router]);

  const logs = useFlowLogEntries({ commandId: useCommandId() });

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>Manage Command | Kite</title>
      </Head>
      {cmd && (
        <FlowPage
          flowData={cmd.flow_source}
          context="command"
          hasUnsavedChanges={hasUnsavedChanges}
          hasUndeployedChanges={hasUndeployedChanges}
          isDeploying={false}
          onDeploy={() => setDeployDialogOpen(true)}
          onChange={onChange}
          isSaving={updateMutation.isPending}
          onSave={save}
          onExit={exit}
          logs={logs}
        />
      )}

      <CommandDeployDialog
        open={deployDialogOpen}
        onOpenChange={setDeployDialogOpen}
      />
    </div>
  );
}
