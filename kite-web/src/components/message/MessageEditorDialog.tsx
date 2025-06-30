import MessageEditor from "@/components/message/MessageEditor";
import MessageEditorPreview from "@/components/message/MessageEditorPreview";
import MessageNav from "@/components/message/MessageNav";
import { Button } from "@/components/ui/button";
import { Drawer, DrawerContent, DrawerTrigger } from "@/components/ui/drawer";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useMessageUpdateMutation } from "@/lib/api/mutations";
import { useMessageQuery } from "@/lib/api/queries";
import { useMessage, useResponseData } from "@/lib/hooks/api";
import { useAppId, useMessageId } from "@/lib/hooks/params";
import { parseMessageData } from "@/lib/message/schemaRestore";
import {
  CurrentMessageStoreProvider,
  useCurrentFlowStore,
  useCurrentMessageStore,
} from "@/lib/message/state";
import { ViewIcon } from "lucide-react";
import Head from "next/head";
import {
  ComponentProps,
  ReactNode,
  useCallback,
  useEffect,
  useRef,
} from "react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";

function MessageEditorDialogInner({
  children,
  onClose,
  messageId,
}: {
  children: ReactNode;
  onClose: () => void;
  messageId: string;
}) {
  const ignoreChange = useRef(false);

  const messageQuery = useMessageQuery(useAppId(), messageId);
  const message = useResponseData(messageQuery, (res) => {
    if (!res.success) {
      toast.error(
        `Failed to load message: ${res?.error.message} (${res?.error.code})`
      );
    }
  });

  const messageStore = useCurrentMessageStore();
  const flowStore = useCurrentFlowStore();

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

  const updateMutation = useMessageUpdateMutation(useAppId(), messageId);

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (open || !message) return;

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
              onClose();
            } else {
              toast.error(
                `Failed to update message: ${res.error.message} (${res.error.code})`
              );
            }
          },
        }
      );
    },
    [message]
  );

  return (
    <Dialog onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="h-[90dvh] w-full max-w-[90dvw] xl:max-w-7xl p-0 !animate-none">
        <DialogTitle className="hidden">Message Editor</DialogTitle>
        <DialogDescription className="hidden">
          Edit your message.
        </DialogDescription>

        <div className="flex min-h-full w-full flex-col">
          <div className="h-full w-full flex flex-col">
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
      </DialogContent>
    </Dialog>
  );
}

export default function MessageEditorDialog(
  props: ComponentProps<typeof MessageEditorDialogInner>
) {
  return (
    <CurrentMessageStoreProvider>
      <MessageEditorDialogInner {...props} />
    </CurrentMessageStoreProvider>
  );
}
