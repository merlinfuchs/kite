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
import { useMessageInstances } from "@/lib/hooks/api";
import GuildSelect from "../common/GuildSelect";
import ChannelSelect from "../common/ChannelSelect";
import MessageSendInstanceEntry from "./MessageSendInstanceEntry";
import { useMessageInstanceCreateMutation } from "@/lib/api/mutations";
import { useAppId, useMessageId } from "@/lib/hooks/params";
import { toast } from "sonner";

export default function MessageSendDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [guildId, setGuildId] = useState("");
  const [channelId, setChannelId] = useState("");

  const instances = useMessageInstances();
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
  }, [setOpen, createMutation, channelId, guildId]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Send Message</DialogTitle>
          <DialogDescription>
            Send the message to the selected channel. The bot must be in the
            server and have the "Manage Webhooks" permission.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col sm:flex-row gap-2">
          <GuildSelect value={guildId} onChange={setGuildId} />
          <ChannelSelect value={channelId} onChange={setChannelId} />
          <LoadingButton
            onClick={createInstance}
            loading={createMutation.isPending}
          >
            Send
          </LoadingButton>
        </div>

        <Separator />
        <div className="flex flex-col space-y-5 max-h-64 overflow-y-auto">
          {instances?.map((instance) => (
            <MessageSendInstanceEntry key={instance!.id} instance={instance!} />
          ))}
          {instances?.length === 0 && (
            <div className="text-muted-foreground text-center text-sm font-light">
              There are no instances of this message yet.
            </div>
          )}
        </div>
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
