import MessagePreviewBox from "@/components/message/MessagePreviewBox";
import { parsePreviewMessage } from "@/lib/message/schema";
import { useMemo } from "react";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";

// FlowMessageViewDialog shows the message a block sends in previews, where
// the message editor can't open: its own message, or its message template
// with the template's name.
export default function FlowMessageViewDialog({
  name,
  message,
}: {
  name?: string;
  message: unknown;
}) {
  const msg = useMemo(() => parsePreviewMessage(message), [message]);
  if (!msg) return null;

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full">
          View Message
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90dvh] overflow-y-auto">
        <DialogTitle>{name ?? "Message"}</DialogTitle>
        <DialogDescription>The message this block sends.</DialogDescription>
        <MessagePreviewBox msg={msg} />
      </DialogContent>
    </Dialog>
  );
}
