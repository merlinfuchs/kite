import { useState } from "react";
import ListSearchInput, { ListSearchEmpty } from "./ListSearchInput";
import { matchesSearch } from "@/lib/search";
import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useMessages } from "@/lib/hooks/api";
import MessageListEntry from "./MessageListEntry";
import MessageCreateDialog from "./MessageCreateDialog";
import { MailPlusIcon } from "lucide-react";
import FlowImportDialog from "./FlowImportDialog";

export default function MessageList() {
  const messages = useMessages();
  const [search, setSearch] = useState("");

  const filtered = messages?.filter((message) =>
    matchesSearch(search, [message!.name, message!.description])
  );

  const messageCreateButton = (
    <div className="flex gap-5 flex-col md:flex-row">
      <MessageCreateDialog>
        <Button>Create message</Button>
      </MessageCreateDialog>
      <FlowImportDialog kind="message">
        <Button variant="outline">Import message</Button>
      </FlowImportDialog>
    </div>
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
          action={messageCreateButton}
        />
      ) : (
        <>
          <div className="flex">{messageCreateButton}</div>

          <ListSearchInput
            value={search}
            onChange={setSearch}
            placeholder="Search message templates"
          />
          {filtered!.length === 0 ? (
            <ListSearchEmpty query={search} />
          ) : (
            filtered!.map((message) => (
              <MessageListEntry message={message!} key={message!.id} />
            ))
          )}
        </>
      )}
    </AutoAnimate>
  );
}
