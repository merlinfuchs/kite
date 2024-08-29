import MessageEditor from "@/components/message/MessageEditor";
import { Button } from "@/components/ui/button";
import { SendIcon, ViewIcon } from "lucide-react";
import MessageEditorPreview from "@/components/message/MessageEditorPreview";
import { Dialog, DialogTrigger } from "@/components/ui/dialog";
import WebhookExecuteDialog from "@/tools/message-creator/components/WebhookExecuteDialog";
import HomeLayout from "@/components/home/HomeLayout";
import { Drawer, DrawerContent, DrawerTrigger } from "@/components/ui/drawer";
import { CurrentMessageStoreProvider } from "@/lib/message/state";

export default function MessageCreatorPage() {
  return (
    <HomeLayout title="Message Creator">
      <CurrentMessageStoreProvider>
        <div className="flex flex-col xl:flex-row h-full">
          <div className="flex flex-col xl:w-7/12 py-8 space-y-8 h-full overflow-y-auto px-3 md:px-5 lg:px-10 no-scrollbar">
            <div className="flex flex-col space-y-5 md:flex-row md:space-y-0 justify-between">
              <div className="flex flex-col space-y-1.5">
                <h1 className="text-2xl font-semibold leading-none tracking-tight">
                  Message Creator
                </h1>
                <p className="text-sm text-muted-foreground">
                  Create good looking Discord messages and send them through
                  webhooks!
                </p>
              </div>
              <Dialog>
                <DialogTrigger asChild>
                  <Button className="flex items-center space-x-2">
                    <SendIcon />
                    <div>Send Message</div>
                  </Button>
                </DialogTrigger>
                <WebhookExecuteDialog />
              </Dialog>
            </div>

            <MessageEditor />
          </div>
          <div className="hidden xl:block py-5 w-5/12 h-full overflow-y-auto pr-5 no-scrollbar">
            <MessageEditorPreview className="rounded-lg" />
          </div>

          <Drawer>
            <DrawerTrigger asChild>
              <Button size="icon" className="fixed bottom-5 right-5 xl:hidden">
                <ViewIcon />
              </Button>
            </DrawerTrigger>
            <DrawerContent>
              <div className="max-h-[80dvh] overlfow-x-hidden overflow-y-auto mt-3">
                <MessageEditorPreview reducePadding />
              </div>
            </DrawerContent>
          </Drawer>
        </div>
      </CurrentMessageStoreProvider>
    </HomeLayout>
  );
}
