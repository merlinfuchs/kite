import MessageEditor from "@/components/message/MessageEditor";
import { Button } from "@/components/ui/button";
import { SendIcon } from "lucide-react";
import MessageEditorPreview from "@/components/message/MessageEditorPreview";
import { Dialog, DialogTrigger } from "@/components/ui/dialog";
import WebhookExecuteDialog from "@/tools/message-creator/components/WebhookExecuteDialog";
import HomeLayout from "@/components/home/HomeLayout";
import MessagePreviewDrawer from "@/components/message/MessagePreviewDrawer";
import { CurrentMessageStoreProvider } from "@/lib/message/state";

export default function MessageCreatorPage() {
  return (
    <HomeLayout title="Message Creator">
      <CurrentMessageStoreProvider>
        <div className="flex flex-col xl:flex-row h-full">
          <div className="flex flex-col xl:w-7/12 pt-8 pb-24 xl:pb-8 space-y-8 h-full overflow-y-auto px-3 md:px-5 lg:px-10 no-scrollbar">
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

            {/* Webhooks can't send interactive components, so there are no flows to edit. */}
            <MessageEditor disableFlowEditor />
          </div>
          <div className="hidden xl:block py-5 w-5/12 h-full overflow-y-auto pr-5 no-scrollbar">
            <MessageEditorPreview className="rounded-lg" />
          </div>

          <MessagePreviewDrawer />
        </div>
      </CurrentMessageStoreProvider>
    </HomeLayout>
  );
}
