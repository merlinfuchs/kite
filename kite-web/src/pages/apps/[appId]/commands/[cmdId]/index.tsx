import FlowPage from "@/components/flow/FlowPage";
import { track } from "@/lib/analytics";
import { useCommandUpdateMutation } from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/dataSchema";
import { useCommand, useFlowLogEntries } from "@/lib/hooks/api";
import { useUnsavedChangesWarning } from "@/lib/hooks/exit";
import { useAppId, useCommandId } from "@/lib/hooks/params";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { isNoteNode } from "@/lib/flow/notes";

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
  // Counts edits, so edits made while a save is running stay unsaved.
  const changeCount = useRef(0);

  const onChange = useCallback(() => {
    changeCount.current++;
    setHasUnsavedChanges(true);
  }, [setHasUnsavedChanges]);

  const save = useCallback(
    async (data: FlowData, editorSession?: string) => {
      const savedChangeCount = changeCount.current;

      try {
        const res = await updateCommand({
          flow_source: data,
          editor_session: editorSession,
        });
        if (!res.success) {
          track("flow_save_failed", {
            kind: "command",
            code: res.error.code,
          });
          toast.error(
            `Failed to update command: ${res.error.message} (${res.error.code})`
          );
          return false;
        }

        track("command_saved", {
          blocks: data.nodes.filter((n) => !isNoteNode(n)).length,
        });
        toast.success("Command saved!");
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

  const leave = useUnsavedChangesWarning(hasUnsavedChanges);

  const exit = useCallback(() => {
    leave({
      pathname: "/apps/[appId]/commands",
      query: { appId: router.query.appId },
    });
  }, [leave, router]);

  const logs = useFlowLogEntries({ commandId: cmdId });
  const versionTarget = useMemo(
    () => ({ appId, type: "command" as const, id: cmdId }),
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
          viewportKey={`command:${cmd.id}`}
          hasUnsavedChanges={hasUnsavedChanges}
          command={cmd}
          bundleInstallId={cmd.bundle_install_id}
          onChange={onChange}
          isSaving={updateMutation.isPending}
          onSave={save}
          onExit={exit}
          logs={logs}
          versionTarget={versionTarget}
        />
      )}
    </div>
  );
}
