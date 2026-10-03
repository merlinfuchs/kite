import FlowPage from "@/components/flow/FlowPage";
import { useEventListenerUpdateMutation } from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/dataSchema";
import { useEventListener, useFlowLogEntries } from "@/lib/hooks/api";
import { useAppId, useEventId } from "@/lib/hooks/params";
import { useUnsavedChangesWarning } from "@/lib/hooks/exit";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useState } from "react";
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

  const updateMutation = useEventListenerUpdateMutation(
    useAppId(),
    useEventId()
  );

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [logsOpen, setLogsOpen] = useState(false);

  const onChange = useCallback(() => {
    setHasUnsavedChanges(true);
  }, [setHasUnsavedChanges]);

  const save = useCallback(
    (data: FlowData, options?: { onSuccess?: () => void }) => {
      setIsSaving(true);

      updateMutation.mutate(
        {
          flow_source: data,
        },
        {
          onSuccess(res) {
            if (res.success) {
              toast.success(
                "Event listener saved! It may take up to a minute for all changes to take effect."
              );
              options?.onSuccess?.();
            } else {
              toast.error(
                `Failed to update event listener: ${res.error.message} (${res.error.code})`
              );
            }
          },
          onSettled() {
            setIsSaving(false);
            setHasUnsavedChanges(false);
          },
        }
      );
    },
    [setIsSaving, setHasUnsavedChanges, updateMutation]
  );

  const exit = useCallback(() => {
    router.push({
      pathname: "/apps/[appId]/events",
      query: { appId: router.query.appId },
    });
  }, [router]);

  useUnsavedChangesWarning(hasUnsavedChanges);

  const logs = useFlowLogEntries({ eventId: useEventId() });

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
          isSaving={isSaving}
          onSave={save}
          onExit={exit}
          logs={logs}
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
