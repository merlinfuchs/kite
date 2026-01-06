import { CommandDeployDialog } from "@/components/app/CommandDeployDialog";
import FlowPage from "@/components/flow/FlowPage";
import {
  useCommandsDeployMutation,
  useCommandUpdateMutation,
} from "@/lib/api/mutations";
import { useLogEntriesQuery } from "@/lib/api/queries";
import { FlowData } from "@/lib/flow/dataSchema";
import { useCommand, useResponseData } from "@/lib/hooks/api";
import { useBeforePageExit } from "@/lib/hooks/exit";
import { useAppId, useCommandId } from "@/lib/hooks/params";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useMemo, useRef, useState } from "react";
import { toast } from "sonner";

export default function AppCommandPage() {
  const ignoreChange = useRef(false);

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
    } else {
      // This is a workaround to ignore the initial change event
      ignoreChange.current = true;
      setTimeout(() => {
        ignoreChange.current = false;
      }, 100);
    }
  });

  const updateMutation = useCommandUpdateMutation(useAppId(), useCommandId());

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [deployDialogOpen, setDeployDialogOpen] = useState(false);

  const onChange = useCallback(() => {
    if (!ignoreChange.current) {
      setHasUnsavedChanges(true);
    }
  }, [setHasUnsavedChanges, ignoreChange]);

  const save = useCallback(
    (data: FlowData) => {
      updateMutation.mutate(
        {
          flow_source: data,
          enabled: true,
        },
        {
          onSuccess(res) {
            if (res.success) {
              toast.success(
                "Command saved! Make sure to deploy the command for the changes to take effect in Discord."
              );
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

  const logsQuery = useLogEntriesQuery(useAppId(), {
    limit: 10,
    commandId: useCommandId(),
    refetchInterval: 10000,
  });
  const logs = useResponseData(logsQuery);

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
