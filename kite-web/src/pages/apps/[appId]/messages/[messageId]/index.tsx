import MessageEditor from "@/components/message/MessageEditor";
import MessageEditorPreview from "@/components/message/MessageEditorPreview";
import MessageNav from "@/components/message/MessageNav";
import MessagePreviewDrawer from "@/components/message/MessagePreviewDrawer";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useMessageUpdateMutation } from "@/lib/api/mutations";
import { track } from "@/lib/analytics";
import { useMessage, useMessageInstances } from "@/lib/hooks/api";
import UnsavedChangesDialog from "@/components/common/UnsavedChangesDialog";
import { useUnsavedChangesWarning } from "@/lib/hooks/exit";
import { useAppId, useMessageId } from "@/lib/hooks/params";
import { messageSchema, parseMessageData } from "@/lib/message/schemaRestore";
import {
  CurrentMessageStoreProvider,
  getMessage,
  useCurrentFlowStore,
  useDocumentStoreApi,
} from "@/lib/message/state";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import {
  useEditorSession,
  useVersionHistory,
} from "@/lib/hooks/versionHistory";
import VersionHistory from "@/components/common/VersionHistory";
import { FlowData } from "@/lib/types/flow.gen";

function AppMessagePageInner() {
  const router = useRouter();

  const ignoreChange = useRef(false);

  const message = useMessage((res) => {
    if (!res.success) {
      toast.error(
        `Failed to load command: ${res?.error.message} (${res?.error.code})`
      );
      if (res.error.code === "unknown_message") {
        router.push({
          pathname: "/apps/[appId]/commands",
          query: { appId: router.query.appId },
        });
      }
    }
  });

  const messageStore = useDocumentStoreApi();
  const flowStore = useCurrentFlowStore();

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  // Puts a message into the editor, and says whether its data was valid.
  const applyMessage = useCallback(
    (data: unknown, flowSources: Record<string, FlowData>) => {
      try {
        messageStore.getState().replaceAll(parseMessageData(data));
        flowStore.getState().replaceAll(flowSources);
        return true;
      } catch (e) {
        toast.error(`Failed to parse message data: ${e}`);
        return false;
      }
    },
    [messageStore, flowStore]
  );

  useEffect(() => {
    if (!message) return;

    ignoreChange.current = true;
    applyMessage(message.data, message.flow_sources);
    messageStore.temporal.getState().clear();
    ignoreChange.current = false;
  }, [message, messageStore, applyMessage]);

  const instances = useMessageInstances();
  const hasInstances = !!instances?.length;

  const appId = useAppId();
  const messageId = useMessageId();
  const updateMutation = useMessageUpdateMutation(appId, messageId);

  const [unsavedDialogOpen, setUnsavedDialogOpen] = useState(false);

  const editorSession = useEditorSession();
  // Set by a restore until it's saved, so it's saved without the session even
  // if the first try fails.
  const restorePending = useRef(false);

  const save = useCallback(
    (options?: { onSuccess?: () => void }) => {
      if (!message) return;

      setIsSaving(true);

      const data = getMessage(messageStore);
      const flowSources = flowStore.getState().flowSources;

      updateMutation.mutate(
        {
          name: message.name,
          description: message.description,
          data: data,
          flow_sources: flowSources,
          editor_session: restorePending.current ? undefined : editorSession,
        },
        {
          onSuccess(res) {
            if (res.success) {
              restorePending.current = false;
              track("message_saved", {
                embeds: data.embeds.length,
                flows: Object.keys(flowSources).length,
              });
              setHasUnsavedChanges(false);
              toast.success("Message saved!", {
                description: hasInstances
                  ? "Messages you already sent keep the old version until you update them in the Send Message dialog."
                  : undefined,
              });
              options?.onSuccess?.();
            } else {
              track("flow_save_failed", {
                kind: "message",
                code: res.error.code,
              });
              toast.error(
                `Failed to update message: ${res.error.message} (${res.error.code})`
              );
            }
          },
          onSettled() {
            setIsSaving(false);
          },
        }
      );
    },
    [
      message,
      hasInstances,
      editorSession,
      updateMutation,
      setIsSaving,
      setHasUnsavedChanges,
      messageStore,
      flowStore,
    ]
  );

  const versionTarget = useMemo(
    () => ({ appId, type: "message" as const, id: messageId }),
    [appId, messageId]
  );
  const history = useVersionHistory({
    target: versionTarget,
    hasUnsavedChanges,
    apply: useCallback(
      (content: Record<string, any>) => {
        if (applyMessage(content.data, content.flow_sources ?? {})) {
          restorePending.current = true;
          save();
        }
      },
      [applyMessage, save]
    ),
  });
  const { resetUndoChain } = history;

  useEffect(() => {
    const onEdit = () => {
      if (ignoreChange.current) return;
      setHasUnsavedChanges(true);
      resetUndoChain();
    };
    const unsubscribeMessage = messageStore.subscribe(onEdit);
    const unsubscribeFlow = flowStore.subscribe(onEdit);

    return () => {
      unsubscribeMessage();
      unsubscribeFlow();
    };
  }, [messageStore, flowStore, resetUndoChain]);

  const leave = useUnsavedChangesWarning(hasUnsavedChanges);

  const doExit = useCallback(() => {
    leave({
      pathname: "/apps/[appId]/messages",
      query: { appId: router.query.appId },
    });
  }, [leave, router]);

  const handleExitRequest = useCallback(() => {
    if (hasUnsavedChanges) {
      setUnsavedDialogOpen(true);
    } else {
      doExit();
    }
  }, [hasUnsavedChanges, doExit]);

  const handleDiscardAndExit = useCallback(() => {
    setUnsavedDialogOpen(false);
    doExit();
  }, [doExit]);

  const handleSaveAndExit = useCallback(() => {
    save({
      onSuccess: () => {
        setUnsavedDialogOpen(false);
        doExit();
      },
    });
  }, [save, doExit]);

  return (
    <div className="flex min-h-[100dvh] w-full flex-col">
      <Head>
        <title>Manage Message | Kite</title>
      </Head>
      <div className="h-[100dvh] w-full flex flex-col">
        <div className="flex-none">
          <MessageNav
            hasUnsavedChanges={hasUnsavedChanges}
            isSaving={isSaving}
            bundleInstallId={message?.bundle_install_id}
            onSave={save}
            onExit={handleExitRequest}
            history={<VersionHistory history={history} isSaving={isSaving} />}
          />
        </div>
        {message && (
          <>
            <div className="flex flex-auto overflow-y-hidden flex-col xl:flex-row h-full">
              <ScrollArea className="flex flex-col xl:w-7/12 pt-3 pb-24 xl:pb-8 space-y-8 h-full px-3 md:px-5 lg:px-10">
                <MessageEditor />
              </ScrollArea>
              <div className="hidden xl:block py-5 w-5/12 h-full pr-5">
                <MessageEditorPreview className="rounded-lg" />
              </div>
            </div>

            <MessagePreviewDrawer />
          </>
        )}
      </div>

      <UnsavedChangesDialog
        open={unsavedDialogOpen}
        onOpenChange={setUnsavedDialogOpen}
        onSaveAndExit={handleSaveAndExit}
        onDiscardAndExit={handleDiscardAndExit}
        isSaving={isSaving}
      />
    </div>
  );
}

export default function AppMessagePage() {
  return (
    <CurrentMessageStoreProvider>
      <AppMessagePageInner />
    </CurrentMessageStoreProvider>
  );
}
