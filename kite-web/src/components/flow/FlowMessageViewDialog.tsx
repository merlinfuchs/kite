import MessagePreviewBox from "@/components/message/MessagePreviewBox";
import { useFlowMessages } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { messageSchema } from "@/lib/message/schema";
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
// the message editor can't open: its own message, or its message template.
export default function FlowMessageViewDialog({ data }: { data: NodeData }) {
  const messages = useFlowMessages();
  const template = data.message_template_id
    ? messages?.find((m) => m.id === data.message_template_id)
    : undefined;
  const source = data.message_template_id ? template?.data : data.message_data;

  const message = useMemo(() => {
    const res = messageSchema.safeParse(source);
    return res.success ? res.data : null;
  }, [source]);

  if (!message) return null;

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full">
          View Message
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90dvh] overflow-y-auto">
        <DialogTitle>{template?.name ?? "Message"}</DialogTitle>
        <DialogDescription>The message this block sends.</DialogDescription>
        <MessagePreviewBox msg={message} />
      </DialogContent>
    </Dialog>
  );
}
