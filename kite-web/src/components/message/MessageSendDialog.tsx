import { ReactNode, useCallback, useState } from "react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import LoadingButton from "../common/LoadingButton";
import { Separator } from "../ui/separator";
import { useMessageFlowInstances, useMessageInstances } from "@/lib/hooks/api";
import GuildSelect from "../common/GuildSelect";
import ChannelSelect from "../common/ChannelSelect";
import MessageSendInstanceEntry from "./MessageSendInstanceEntry";
import { useMessageInstanceCreateMutation } from "@/lib/api/mutations";
import { useAppId, useMessageId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { ScrollArea } from "../ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";

export default function MessageSendDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [guildId, setGuildId] = useState<string | null>(null);
  const [channelId, setChannelId] = useState<string | null>(null);

  const [tab, setTab] = useState<"manual" | "flow">("manual");

  const manualInstances = useMessageInstances();
  const flowInstances = useMessageFlowInstances(open && tab === "flow");
  const instances = tab === "flow" ? flowInstances : manualInstances;
  const createMutation = useMessageInstanceCreateMutation(
    useAppId(),
    useMessageId()
  );

  const createInstance = useCallback(() => {
    if (createMutation.isPending || !guildId || !channelId) return;

    createMutation.mutate(
      {
        discord_guild_id: guildId,
        discord_channel_id: channelId,
      },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("Message sent!");
          } else {
            toast.error(
              `Failed to send message: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }, [createMutation, channelId, guildId]);

  // We had to set modal={false} because otherwise the popovers don't work
  return (
    <Dialog open={open} onOpenChange={setOpen} modal={false}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Send Message</DialogTitle>
          <DialogDescription>
            Send the message to the selected channel. The bot must be in the
            server and have the &quot;Manage Webhooks&quot; permission.
          </DialogDescription>
          <DialogDescription>
            Messages you already sent don&apos;t change when you save the
            template. Click Update to apply your saved changes, including button
            flows, to them.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col sm:flex-row gap-2 overflow-x-hidden">
          <GuildSelect value={guildId} onChange={setGuildId} />
          <ChannelSelect
            guildId={guildId || null}
            value={channelId}
            onChange={setChannelId}
          />
          <LoadingButton
            onClick={createInstance}
            loading={createMutation.isPending}
          >
            Send
          </LoadingButton>
        </div>

        <Separator />
        <Tabs value={tab} onValueChange={(v) => setTab(v as "manual" | "flow")}>
          <TabsList className="w-full">
            <TabsTrigger value="manual" className="flex-1">
              Sent from here
            </TabsTrigger>
            <TabsTrigger value="flow" className="flex-1">
              Sent by flows
            </TabsTrigger>
          </TabsList>
        </Tabs>
        <ScrollArea className="overflow-y-hidden max-h-64 pr-3">
          <div className="flex flex-col space-y-5">
            {instances?.map((instance) => (
              <MessageSendInstanceEntry
                key={instance!.id}
                instance={instance!}
              />
            ))}
            {tab === "flow" && instances?.length === 100 && (
              <div className="text-muted-foreground text-center text-sm font-light">
                Only the 100 newest messages are shown.
              </div>
            )}
            {instances?.length === 0 && (
              <div className="text-muted-foreground text-center text-sm font-light">
                {tab === "flow"
                  ? "No flow has sent this message yet. Ephemeral messages aren't listed because they can't be updated."
                  : "There are no instances of this message yet."}
              </div>
            )}
          </div>
        </ScrollArea>
        <Separator />

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
