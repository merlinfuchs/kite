import FlowPage from "@/components/flow/FlowPage";
import { useEventListenerUpdateMutation } from "@/lib/api/mutations";
import { FlowData } from "@/lib/flow/data";
import { useEventListener } from "@/lib/hooks/api";
import { useAppId, useEventId } from "@/lib/hooks/params";
import { useBeforePageExit } from "@/lib/hooks/exit";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import { LogEntryListDrawer } from "@/components/app/LogEntryListDrawer";

export default function AppEventListenerPage() {
  const ignoreChange = useRef(false);

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
    } else {
      // This is a workaround to ignore the initial change event
      ignoreChange.current = true;
      setTimeout(() => {
        ignoreChange.current = false;
      }, 100);
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
    if (!ignoreChange.current) {
      setHasUnsavedChanges(true);
    }
  }, [setHasUnsavedChanges, ignoreChange]);

  const save = useCallback(
    (data: FlowData) => {
      setIsSaving(true);

      updateMutation.mutate(
        {
          flow_source: data,
          enabled: true,
        },
        {
          onSuccess(res) {
            if (res.success) {
              toast.success(
                "Event listener saved! It may take up to a minute for all changes to take effect."
              );
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

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>Manage Event Listener | Kite</title>
      </Head>
      {listener && (
        <FlowPage
          flowData={listener.flow_source}
          context="event_discord"
          hasUnsavedChanges={hasUnsavedChanges}
          onChange={onChange}
          isSaving={isSaving}
          onSave={save}
          onExit={exit}
          onLogsView={() => setLogsOpen(true)}
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
