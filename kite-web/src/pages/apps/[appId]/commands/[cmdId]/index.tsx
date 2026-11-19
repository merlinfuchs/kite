import FlowPage from "@/components/flow/FlowPage";
import { track } from "@/lib/analytics";
import { useCommandUpdateMutation } from "@/lib/api/mutations";
import { useAppQuery, useCommandQuery } from "@/lib/api/queries";
import { FlowData } from "@/lib/flow/dataSchema";
import { useFlowLogEntries, useResponseData } from "@/lib/hooks/api";
import { useUnsavedChangesWarning } from "@/lib/hooks/exit";
import { saveErrorMessage, useSaveBase } from "@/lib/hooks/versionHistory";
import { useAppId, useCommandId } from "@/lib/hooks/params";
import { usePwaInstall } from "@/components/common/PwaExperience";
import { BotLoadingGate } from "@/components/app/BotLoadingScreen";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { isNoteNode } from "@/lib/flow/notes";

export default function AppCommandPage() {
  const router = useRouter();
  const appId = useAppId();
  const cmdId = useCommandId();
  const appQuery = useAppQuery(appId);
  const commandQuery = useCommandQuery(appId, cmdId);
  const { isInstalled, navigateWithPwaTransition } = usePwaInstall();
  const cmd = useResponseData(commandQuery, (res) => {
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

  const updateMutation = useCommandUpdateMutation(appId, cmdId);
  const { mutateAsync: updateCommand } = updateMutation;

  const base = useSaveBase(cmd?.version);
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
          base_version: base.current ?? undefined,
        });
        if (!res.success) {
          track("flow_save_failed", {
            kind: "command",
            code: res.error.code,
          });
          if (res.error.code === "item_outdated") base.current = null;
          toast.error(saveErrorMessage("command", res.error));
          return false;
        }

        track("command_saved", {
          blocks: data.nodes.filter((n) => !isNoteNode(n)).length,
        });
        toast.success("Command saved.");
        base.current = res.data.version;
        if (changeCount.current === savedChangeCount) {
          setHasUnsavedChanges(false);
        }
        return true;
      } catch (err) {
        toast.error(`Failed to update command: ${err}`);
        return false;
      }
    },
    [setHasUnsavedChanges, base, updateCommand]
  );

  const leave = useUnsavedChangesWarning(hasUnsavedChanges);

  const exit = useCallback(() => {
    navigateWithPwaTransition(`/apps/${appId}/commands`, () =>
      leave({
        pathname: "/apps/[appId]/commands",
        query: { appId: router.query.appId },
      })
    );
  }, [appId, leave, navigateWithPwaTransition, router]);

  const logs = useFlowLogEntries({ commandId: cmdId });
  const versionTarget = useMemo(
    () => ({ appId, type: "command" as const, id: cmdId }),
    [appId, cmdId]
  );

  return (
    <BotLoadingGate
      enabled={isInstalled}
      isLoading={appQuery.isPending || commandQuery.isPending}
    >
      <div className="flex min-h-[100dvh] w-full flex-col">
        <Head>
          <title>Manage command | Kite</title>
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
    </BotLoadingGate>
  );
}
