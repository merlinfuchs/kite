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

export default function MessageList() {
  const messages = useMessages();
  const [search, setSearch] = useState("");

  const filtered = messages?.filter((message) =>
    matchesSearch(search, [message!.name, message!.description])
  );

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
          <div className="flex">{messageCreateButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}
