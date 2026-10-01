import FlowPage, { FlowSaveOptions } from "@/components/flow/FlowPage";
import { useEventListenerUpdateMutation } from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/dataSchema";
import { useEventListener, useFlowLogEntries } from "@/lib/hooks/api";
import { useAppId, useEventId } from "@/lib/hooks/params";
import { useBeforePageExit } from "@/lib/hooks/exit";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { LogEntryListDrawer } from "@/components/app/LogEntryListDrawer";

export default function AppEventListenerPage() {
  const router = useRouter();
  const listener = useEventListener((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load event listener: ${res?.error.message} (${res?.error.code})`
      );
      if (res.error.code === "unknown_event_listener") {
        router.push({
          pathname: "/apps/[appId]/events",
          query: { appId: router.query.appId },
        });
      }
    }
  });

  const appId = useAppId();
  const eventId = useEventId();
  const updateMutation = useEventListenerUpdateMutation(appId, eventId);
  const { mutateAsync: updateEventListener } = updateMutation;

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [logsOpen, setLogsOpen] = useState(false);
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
        const res = await updateEventListener({
          flow_source: data,
          auto_save: !!options?.auto,
        });
        if (!res.success) {
          toast.error(
            `Failed to update event listener: ${res.error.message} (${res.error.code})`
          );
          return false;
        }

        // Auto-save would show this every few seconds.
        if (!options?.auto) {
          toast.success(
            "Event listener saved! It may take up to a minute for all changes to take effect."
          );
        }
        if (changeCount.current === savedChangeCount) {
          setHasUnsavedChanges(false);
        }
        return true;
      } catch (err) {
        toast.error(`Failed to update event listener: ${err}`);
        return false;
      }
    },
    [setHasUnsavedChanges, updateEventListener]
  );

  const exit = useCallback(() => {
    if (hasUnsavedChanges) {
      if (
        !confirm("You have unsaved changes. Are you sure you want to exit?")
      ) {
        return;
      }
    }

    router.push({
      pathname: "/apps/[appId]/events",
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

  const logs = useFlowLogEntries({ eventId });
  const versionTarget = useMemo(
    () => ({ appId, eventListenerId: eventId }),
    [appId, eventId]
  );

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>Manage Event Listener | Kite</title>
      </Head>
      {listener && (
        <FlowPage
          flowData={listener.flow_source}
          context={
            listener.source === "schedule" ? "event_schedule" : "event_discord"
          }
          hasUnsavedChanges={hasUnsavedChanges}
          onChange={onChange}
          isSaving={updateMutation.isPending}
          onSave={save}
          onExit={exit}
          logs={logs}
          versionTarget={versionTarget}
        />
      )}
      <LogEntryListDrawer
        eventId={listener?.id}
        open={logsOpen}
        onOpenChange={setLogsOpen}
      />
    </div>
  );
}
