import FlowPage, { FlowSaveOptions } from "@/components/flow/FlowPage";
import { track } from "@/lib/analytics";
import { useEventListenerUpdateMutation } from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/dataSchema";
import { useEventListener, useFlowLogEntries } from "@/lib/hooks/api";
import { useAppId, useEventId } from "@/lib/hooks/params";
import { useUnsavedChangesWarning } from "@/lib/hooks/exit";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { LogEntryListDrawer } from "@/components/app/LogEntryListDrawer";
import { capitalize, eventSourceName } from "@/lib/flow/eventTypes";

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
  const name = eventSourceName(listener?.source);

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
          track("flow_save_failed", {
            kind: "event_listener",
            code: res.error.code,
          });
          toast.error(
            `Failed to update ${name.singular}: ${res.error.message} (${res.error.code})`
          );
          return false;
        }

        // Auto-save would show this every few seconds.
        if (!options?.auto) {
          track("event_listener_saved", {
            type: res.data.type,
            blocks: data.nodes.length,
          });
          toast.success(
            `${capitalize(
              name.singular
            )} saved! It may take up to a minute for all changes to take effect.`
          );
        }
        if (changeCount.current === savedChangeCount) {
          setHasUnsavedChanges(false);
        }
        return true;
      } catch (err) {
        toast.error(`Failed to update ${name.singular}: ${err}`);
        return false;
      }
    },
    [setHasUnsavedChanges, updateEventListener, name]
  );

  const leave = useUnsavedChangesWarning(hasUnsavedChanges);

  const exit = useCallback(() => {
    // Back to the list of the event listener's source.
    leave({
      pathname: name.path,
      query: { appId: router.query.appId },
    });
  }, [leave, router, name]);

  const logs = useFlowLogEntries({ eventId });
  const versionTarget = useMemo(
    () => ({ appId, eventListenerId: eventId }),
    [appId, eventId]
  );

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>{`Manage ${name.title} | Kite`}</title>
      </Head>
      {listener && (
        <FlowPage
          flowData={listener.flow_source}
          context={
            listener.source === "schedule"
              ? "event_schedule"
              : listener.source === "webhook"
              ? "event_webhook"
              : "event_discord"
          }
          viewportKey={`event:${listener.id}`}
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
