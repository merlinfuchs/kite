import { CommandDeployDialog } from "@/components/app/CommandDeployDialog";
import FlowPage, { FlowSaveOptions } from "@/components/flow/FlowPage";
import {
  useCommandsDeployMutation,
  useCommandUpdateMutation,
} from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/dataSchema";
import { useCommand, useFlowLogEntries } from "@/lib/hooks/api";
import { useBeforePageExit } from "@/lib/hooks/exit";
import { useAppId, useCommandId } from "@/lib/hooks/params";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useMemo, useRef, useState } from "react";
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

  const appId = useAppId();
  const cmdId = useCommandId();
  const updateMutation = useCommandUpdateMutation(appId, cmdId);
  const { mutateAsync: updateCommand } = updateMutation;

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [deployDialogOpen, setDeployDialogOpen] = useState(false);
  // Counts edits, so edits made while a save is running stay unsaved.
  const changeCount = useRef(0);

  const onChange = useCallback(() => {
    changeCount.current++;
    setHasUnsavedChanges(true);
  }, [setHasUnsavedChanges]);

  const save = useCallback(
    async (data: FlowData, options?: FlowSaveOptions) => {
      const savedChangeCount = changeCount.current;

      try {
        const res = await updateCommand({
          flow_source: data,
          auto_save: !!options?.auto,
        });
        if (!res.success) {
          toast.error(
            `Failed to update command: ${res.error.message} (${res.error.code})`
          );
          return false;
        }

        // Auto-save would show this every few seconds.
        if (!options?.auto) {
          toast.success(
            "Command saved! Make sure to deploy the command for the changes to take effect in Discord."
          );
        }
        if (changeCount.current === savedChangeCount) {
          setHasUnsavedChanges(false);
        }
        return true;
      } catch (err) {
        toast.error(`Failed to update command: ${err}`);
        return false;
      }
    },
    [setHasUnsavedChanges, updateCommand]
  );

  const hasUndeployedChanges = useMemo(() => {
    return (
      cmd && new Date(cmd!.updated_at) > new Date(cmd!.last_deployed_at || 0)
    );
  }, [cmd]);

  const exit = useCallback(() => {
    if (hasUnsavedChanges) {
      if (
        !confirm("You have unsaved changes. Are you sure you want to exit?")
      ) {
        return;
      }
    }

    router.push({
      pathname: "/apps/[appId]/commands",
      query: { appId: router.query.appId },
    });
  }, [hasUnsavedChanges, router]);

  useBeforePageExit(
    (e) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
        return "You have unsaved changes. Are you sure you want to exit?";
      }
    },
    [hasUnsavedChanges]
  );

  const logs = useFlowLogEntries({ commandId: cmdId });
  const versionTarget = useMemo(
    () => ({ appId, commandId: cmdId }),
    [appId, cmdId]
  );

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
          versionTarget={versionTarget}
        />
      )}

      <CommandDeployDialog
        open={deployDialogOpen}
        onOpenChange={setDeployDialogOpen}
      />
    </div>
  );
}
