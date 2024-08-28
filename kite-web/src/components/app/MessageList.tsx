import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useMessages } from "@/lib/hooks/api";
import MessageListEntry from "./MessageListEntry";
import MessageCreateDialog from "./MessageCreateDialog";

export default function MessageList() {
  const messages = useMessages();

  const messageCreateButton = (
    <MessageCreateDialog>
      <Button>Create message</Button>
    </MessageCreateDialog>
  );

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!messages ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : messages.length === 0 ? (
        <AppEmptyPlaceholder
          title="There are no message templates"
          description="You can start now by creating the first message template!"
          action={messageCreateButton}
        />
      ) : (
        <>
          {messages.map((message, i) => (
            <MessageListEntry message={message!} key={i} />
          ))}
          <div className="flex">{messageCreateButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}
