import { ReactNode, useCallback, useMemo, useState } from "react";
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
import {
  useAppStateGuildChannel,
  useMessage,
  useMessageInstances,
} from "@/lib/hooks/api";
import GuildSelect from "../common/GuildSelect";
import ChannelSelect from "../common/ChannelSelect";
import MessageSendInstanceEntry from "./MessageSendInstanceEntry";
import { useMessageInstanceCreateMutation } from "@/lib/api/mutations";
import { useAppId, useMessageId } from "@/lib/hooks/params";
import { toast } from "sonner";
import { ScrollArea } from "../ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "../ui/tabs";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { badgeVariants } from "../ui/badge";
import { cn } from "@/lib/utils";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert";
import { AlertCircleIcon, HashIcon, MessagesSquareIcon } from "lucide-react";

// Mirrors maxFlowInstances in the API.
const maxFlowInstances = 100;

// Forum and media channels only contain posts, so sending creates a new post.
const forumChannelTypes = [15, 16];
// Discord limits for forum posts.
const maxPostTitleLength = 100;
const maxPostTags = 5;

type Tab = "manual" | "flow";
type SendMode = "message" | "forum";

export default function MessageSendDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [guildId, setGuildId] = useState<string | null>(null);
  const [channelId, setChannelId] = useState<string | null>(null);
  const [mode, setMode] = useState<SendMode>("message");
  const [postTitle, setPostTitle] = useState("");
  const [postTags, setPostTags] = useState<string[]>([]);

  const channel = useAppStateGuildChannel(guildId, channelId);
  const isForum = mode === "forum";
  const availableTags = channel?.available_tags ?? [];

  // Tags belong to a channel, so they're cleared when the channel changes.
  const selectChannel = useCallback((id: string | null) => {
    setChannelId(id);
    setPostTags([]);
  }, []);

  // The tag buttons are disabled once maxPostTags are selected.
  const toggleTag = useCallback((tagId: string) => {
    setPostTags((tags) =>
      tags.includes(tagId) ? tags.filter((t) => t !== tagId) : [...tags, tagId]
    );
  }, []);

  // Only flows fill in placeholders. Sent from here they stay as they are,
  // and Discord rejects them in URLs.
  const message = useMessage();
  const hasPlaceholders = useMemo(
    () => !!message && JSON.stringify(message.data).includes("{{"),
    [message]
  );

  const [tab, setTab] = useState<Tab>("manual");

  const manualInstances = useMessageInstances();
  const flowInstances = useMessageInstances({
    sentBy: "flow",
    enabled: open && tab === "flow",
  });
  const instances = tab === "flow" ? flowInstances : manualInstances;
  const createMutation = useMessageInstanceCreateMutation(
    useAppId(),
    useMessageId()
  );

  const createInstance = useCallback(() => {
    if (createMutation.isPending || !guildId || !channelId) return;

    const title = postTitle.trim();
    if (isForum && !title) {
      toast.error("Enter a title for the forum post.");
      return;
    }
    if (isForum && channel?.require_tag && postTags.length === 0) {
      toast.error("This forum requires at least one tag.");
      return;
    }

    createMutation.mutate(
      {
        discord_guild_id: guildId,
        discord_channel_id: channelId,
        ...(isForum ? { thread_name: title, applied_tags: postTags } : {}),
      },
      {
        onSuccess(res) {
          if (res.success) {
            if (isForum) {
              toast.success("Forum post created!");
              setPostTitle("");
              setPostTags([]);
            } else {
              toast.success("Message sent!");
            }
          } else {
            toast.error(
              `Failed to send message: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }, [
    createMutation,
    channelId,
    guildId,
    channel,
    isForum,
    postTitle,
    postTags,
  ]);

  // We had to set modal={false} because otherwise the popovers don't work
  return (
    <Dialog open={open} onOpenChange={setOpen} modal={false}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Send Message</DialogTitle>
          <DialogDescription>
            Send the message to the selected channel. The bot must be in the
            server and have the &quot;Manage Webhooks&quot; permission. Switch
            to Forum Post to start a new post in a forum or media channel.
          </DialogDescription>
          <DialogDescription>
            Messages you already sent don&apos;t change when you save the
            template. Click Update to apply your saved changes, including button
            flows, to them.
          </DialogDescription>
        </DialogHeader>

        {hasPlaceholders && (
          <Alert variant="destructive">
            <AlertCircleIcon className="h-4 w-4" />
            <AlertTitle>Placeholders aren&apos;t filled in here</AlertTitle>
            <AlertDescription>
              Placeholders like {"{{user.username}}"} only work when a flow
              sends the message, for example with the Create channel message
              block in an event listener. Sent from here, they show up as
              written and placeholders in URLs make sending fail.
            </AlertDescription>
          </Alert>
        )}

        <Tabs
          value={mode}
          onValueChange={(v) => {
            setMode(v as SendMode);
            // The channel lists differ per mode, so the selection doesn't carry over.
            selectChannel(null);
          }}
        >
          <TabsList className="w-full">
            <TabsTrigger value="message" className="flex-1 gap-1.5">
              <HashIcon className="h-4 w-4" />
              Channel Message
            </TabsTrigger>
            <TabsTrigger value="forum" className="flex-1 gap-1.5">
              <MessagesSquareIcon className="h-4 w-4" />
              Forum Post
            </TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="flex flex-col sm:flex-row gap-2 overflow-x-hidden">
          <GuildSelect value={guildId} onChange={setGuildId} />
          <ChannelSelect
            guildId={guildId || null}
            value={channelId}
            onChange={selectChannel}
            types={isForum ? forumChannelTypes : undefined}
            placeholder={isForum ? "Select forum..." : undefined}
            emptyText={isForum ? "No forum channel found." : undefined}
          />
          <LoadingButton
            onClick={createInstance}
            loading={createMutation.isPending}
          >
            {isForum ? "Create Post" : "Send"}
          </LoadingButton>
        </div>

        {isForum && (
          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="forum-post-title">Post Title</Label>
              <Input
                id="forum-post-title"
                value={postTitle}
                onChange={(e) => setPostTitle(e.target.value)}
                maxLength={maxPostTitleLength}
                placeholder="Title of the new post"
              />
            </div>
            {availableTags.length > 0 && (
              <div className="flex flex-col gap-1.5">
                <Label>
                  Tags{" "}
                  <span className="text-muted-foreground font-normal">
                    ({postTags.length}/{maxPostTags})
                  </span>
                </Label>
                <div className="flex flex-wrap gap-1.5">
                  {availableTags.map((tag) => {
                    const selected = postTags.includes(tag.id);
                    return (
                      <button
                        key={tag.id}
                        type="button"
                        onClick={() => toggleTag(tag.id)}
                        disabled={!selected && postTags.length >= maxPostTags}
                        className={cn(
                          badgeVariants({
                            variant: selected ? "default" : "outline",
                          }),
                          "cursor-pointer disabled:cursor-not-allowed disabled:opacity-50"
                        )}
                      >
                        {tag.name}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        )}

        <Separator />
        <Tabs value={tab} onValueChange={(v) => setTab(v as Tab)}>
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
            {tab === "flow" && instances?.length === maxFlowInstances && (
              <div className="text-muted-foreground text-center text-sm font-light">
                Only the {maxFlowInstances} newest messages are shown.
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
