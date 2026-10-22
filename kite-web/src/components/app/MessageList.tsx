import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useMessages } from "@/lib/hooks/api";
import MessageListEntry from "./MessageListEntry";
import MessageCreateDialog from "./MessageCreateDialog";
import { MailPlusIcon } from "lucide-react";
import FlowImportDialog from "./FlowImportDialog";
import SearchableList, { ListActions } from "./SearchableList";

export default function MessageList() {
  const messages = useMessages();

  const messageActions = (
    <>
      <MessageCreateDialog>
        <Button>Create message</Button>
      </MessageCreateDialog>
      <FlowImportDialog kind="message">
        <Button variant="outline">Import message</Button>
      </FlowImportDialog>
    </>
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
          icon={MailPlusIcon}
          title="No message templates yet"
          description="Design a message once and send it from any of your commands and events."
          action={<ListActions>{messageActions}</ListActions>}
        />
      ) : (
        <SearchableList
          items={messages}
          fields={(message) => [message.name, message.description]}
          renderItem={(message) => (
            <MessageListEntry message={message} key={message.id} />
          )}
          placeholder="Search message templates"
          actions={messageActions}
        />
      )}
    </AutoAnimate>
  );
}
