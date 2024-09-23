import MessageEditor from "@/components/message/MessageEditor";
import MessageEditorPreview from "@/components/message/MessageEditorPreview";
import MessageNav from "@/components/message/MessageNav";
import { Button } from "@/components/ui/button";
import { Drawer, DrawerContent, DrawerTrigger } from "@/components/ui/drawer";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useMessageUpdateMutation } from "@/lib/api/mutations";
import { useMessage } from "@/lib/hooks/api";
import { useBeforePageExit } from "@/lib/hooks/exit";
import { useAppId, useMessageId } from "@/lib/hooks/params";
import { messageSchema, parseMessageData } from "@/lib/message/schemaRestore";
import {
  CurrentMessageStoreProvider,
  useCurrentFlowStore,
  useCurrentMessage,
  useCurrentMessageStore,
} from "@/lib/message/state";
import { ViewIcon } from "lucide-react";
import Head from "next/head";
import { useRouter } from "next/router";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

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

  const messageStore = useCurrentMessageStore();
  const flowStore = useCurrentFlowStore();

  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    const unsubscribeMessage = messageStore.subscribe(() => {
      if (!ignoreChange.current) {
        setHasUnsavedChanges(true);
      }
    });

    const unsubscribeFlow = flowStore.subscribe(() => {
      if (!ignoreChange.current) {
        setHasUnsavedChanges(true);
      }
    });

    return () => {
      unsubscribeMessage();
      unsubscribeFlow();
    };
  }, [ignoreChange, messageStore, flowStore]);

  useEffect(() => {
    if (!message) return;

    try {
      const data = parseMessageData(message.data);

      ignoreChange.current = true;
      messageStore.getState().replace(data);
      messageStore.temporal.getState().clear();
      flowStore.getState().replaceAll(message.flow_sources);
      ignoreChange.current = false;
    } catch (e) {
      toast.error(`Failed to parse message data: ${e}`);
    }
  }, [message, messageStore, flowStore]);

  const updateMutation = useMessageUpdateMutation(useAppId(), useMessageId());

  const save = useCallback(() => {
    if (!message) return;

    setIsSaving(true);

    const data = messageStore.getState();
    const flowSources = flowStore.getState().flowSources;

    updateMutation.mutate(
      {
        name: message.name,
        description: message.description,
        data: data,
        flow_sources: flowSources,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Message saved!");
          } else {
            toast.error(
              `Failed to update message: ${res.error.message} (${res.error.code})`
            );
          }
        },
        onSettled() {
          setIsSaving(false);
          setHasUnsavedChanges(false);
        },
      }
    );
  }, [
    message,
    updateMutation,
    setIsSaving,
    setHasUnsavedChanges,
    messageStore,
    flowStore,
  ]);

  const exit = useCallback(() => {
    if (hasUnsavedChanges) {
      if (
        !confirm("You have unsaved changes. Are you sure you want to exit?")
      ) {
        return;
      }
    }

    router.push({
      pathname: "/apps/[appId]/messages",
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
        <title>Manage Message | Kite</title>
      </Head>
      <div className="h-[100dvh] w-[100dvw] flex flex-col">
        <div className="flex-none">
          <MessageNav
            hasUnsavedChanges={hasUnsavedChanges}
            isSaving={isSaving}
            onSave={save}
            onExit={exit}
          />
        </div>
        {message && (
          <>
            <div className="flex flex-auto overflow-y-hidden flex-col xl:flex-row h-full">
              <ScrollArea className="flex flex-col xl:w-7/12 pt-3 pb-8 space-y-8 h-full px-3 md:px-5 lg:px-10">
                <MessageEditor />
              </ScrollArea>
              <div className="hidden xl:block py-5 w-5/12 h-full pr-5">
                <MessageEditorPreview className="rounded-lg" />
              </div>
            </div>

            <Drawer>
              <DrawerTrigger asChild>
                <Button
                  size="icon"
                  className="fixed bottom-5 right-5 xl:hidden"
                >
                  <ViewIcon />
                </Button>
              </DrawerTrigger>
              <DrawerContent>
                <div className="max-h-[80dvh] overlfow-x-hidden overflow-y-auto mt-3">
                  <MessageEditorPreview reducePadding />
                </div>
              </DrawerContent>
            </Drawer>
          </>
        )}
      </div>
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
