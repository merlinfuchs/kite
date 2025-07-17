import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Fragment, ReactNode, useEffect, useRef } from "react";
import { Button } from "../ui/button";
import LogEntryList from "./LogEntryList";
import { useLogEntriesQuery } from "@/lib/api/queries";
import { useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import LogLevelBadge from "./LogLevelBadge";
import { formatRelative } from "date-fns";

export function LogEntryListDrawer({
  open,
  onOpenChange,
  commandId,
  eventId,
  messageId,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  commandId?: string;
  eventId?: string;
  messageId?: string;
}) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle className="sr-only">Logs</DrawerTitle>
          <DrawerDescription className="sr-only">
            View logs for the selected entity.
          </DrawerDescription>
        </DrawerHeader>
        <LogList
          commandId={commandId}
          eventId={eventId}
          messageId={messageId}
        />
        <DrawerFooter>
          <DrawerClose asChild>
            <Button variant="outline" className="w-full">
              Close Logs
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}

function LogList({
  commandId,
  eventId,
  messageId,
}: {
  commandId?: string;
  eventId?: string;
  messageId?: string;
}) {
  const query = useLogEntriesQuery(useAppId(), {
    commandId,
    eventId,
    messageId,
    limit: 10,
  });
  const data = useResponseData(query);

  const mounted = useRef(false);

  useEffect(() => {
    if (query.data && !query.isPending && !mounted.current) {
      query.refetch();
    }
    mounted.current = true;
  }, [query]);

  if (data?.length === 0) {
    return (
      <div className="h-48 flex items-center justify-center">
        <div className="flex flex-col gap-2 text-muted-foreground text-sm">
          No logs yet.
        </div>
      </div>
    );
  }

  return (
    <div className="px-3 pb-3">
      <div className="flex flex-col gap-2">
        {data?.map((entry) => (
          <div
            key={entry!.id}
            className="flex gap-5 items-center bg-muted/50 rounded-md px-5 py-3"
          >
            <div className="w-16 flex-none">
              <LogLevelBadge level={entry!.level} />
            </div>
            <div className="flex-auto break-words font-mono text-sm">
              {entry!.message}
            </div>
            <div className="w-40 flex-none text-left">
              {formatRelative(new Date(entry!.created_at), new Date())}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
