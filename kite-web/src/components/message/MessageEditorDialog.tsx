import MessageEditor from "@/components/message/MessageEditor";
import MessageEditorPreview from "@/components/message/MessageEditorPreview";
import { Button } from "@/components/ui/button";
import { Drawer, DrawerContent, DrawerTrigger } from "@/components/ui/drawer";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useMessageUpdateMutation } from "@/lib/api/mutations";
import { useMessageQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { parseMessageData } from "@/lib/message/schemaRestore";
import {
  CurrentMessageStoreProvider,
  useCurrentFlowStore,
  useCurrentMessageStore,
} from "@/lib/message/state";
import { ViewIcon } from "lucide-react";
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
import { MessageData } from "@/lib/types/message.gen";

function MessageEditorDialogInner({
  children,
  message,
  onClose,
}: {
  children: ReactNode;
  message: MessageData;
  onClose: (message: MessageData) => void;
}) {
  const ignoreChange = useRef(false);

  const messageStore = useCurrentMessageStore();
  const flowStore = useCurrentFlowStore();

  useEffect(() => {
    try {
      const data = parseMessageData(message);

      ignoreChange.current = true;
      messageStore.getState().replace(data);
      messageStore.temporal.getState().clear();
      ignoreChange.current = false;
    } catch (e) {
      toast.error(`Failed to parse message data: ${e}`);
    }
  }, [message, messageStore, flowStore]);

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (open || !message) return;

      const data = messageStore.getState();
      onClose(data);
    },
    [message, onClose, messageStore]
  );

  return (
    <Dialog onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="h-full sm:h-[90dvh] w-full md:max-w-[90dvw] xl:max-w-7xl p-0 !animate-none">
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
                    <MessageEditor disableFlowEditor />
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
