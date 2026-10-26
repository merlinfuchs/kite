import { LogEntry } from "@/lib/types/wire.gen";
import LogLevelBadge from "../app/LogLevelBadge";
import { formatRelative } from "date-fns";
import { ScrollArea } from "../ui/scroll-area";
import { useReactFlow, useStore, useStoreApi } from "@xyflow/react";
import { useCallback } from "react";
import { Button } from "../ui/button";
import { CrosshairIcon } from "lucide-react";

export default function FlowLogList({
  logs,
  hideHeader,
  onShowNode,
}: {
  logs?: LogEntry[];
  hideHeader?: boolean;
  // Called after a block of an error has been selected.
  onShowNode?: () => void;
}) {
  const store = useStoreApi();
  const { fitView } = useReactFlow();

  // Blocks deleted since the error can't be shown.
  const existingNodeIds = useStore(
    useCallback(
      (s) =>
        (logs ?? [])
          .map((entry) => entry.node_id)
          .filter((id): id is string => !!id && s.nodeLookup.has(id))
          .join(" "),
      [logs]
    )
  ).split(" ");

  const showNode = useCallback(
    (id: string) => {
      store.getState().addSelectedNodes([id]);
      fitView({ nodes: [{ id }], duration: 300, maxZoom: 1 });
      onShowNode?.();
    },
    [store, fitView, onShowNode]
  );

  if (!logs?.length) {
    return (
      <div className="h-40 flex items-center justify-center">
        <div className="flex flex-col gap-2 text-muted-foreground text-sm">
          No logs yet.
        </div>
      </div>
    );
  }

  return (
    <div className="w-full h-full flex flex-col min-h-0 overflow-hidden">
      {!hideHeader && (
        <div className="p-5 flex-none">
          <div className="text-xl font-bold text-foreground mb-2">Logs</div>
          <div className="text-muted-foreground">
            View the logs for your app. Some logs are produced by Kite itself,
            but you can also add your own logs to your flows.
          </div>
        </div>
      )}

      <ScrollArea className="flex-1 min-h-0">
        <div className="flex flex-col gap-2 px-4">
          {logs?.map((entry) => (
            <div
              key={entry!.id}
              className="flex flex-col gap-2 bg-muted/40 rounded-sm px-4 py-3"
            >
              <div className="flex items-center justify-between">
                <div className="w-40 flex-none text-left text-sm text-muted-foreground">
                  {formatRelative(new Date(entry!.created_at), new Date())}
                </div>

                <div className="flex-none flex items-center gap-2">
                  {entry.node_id && existingNodeIds.includes(entry.node_id) && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 gap-1.5 px-2 text-muted-foreground"
                      onClick={() => showNode(entry.node_id!)}
                    >
                      <CrosshairIcon className="size-3.5" />
                      Show block
                    </Button>
                  )}
                  <LogLevelBadge level={entry!.level} />
                </div>
              </div>

              <div className="flex-auto break-words font-mono text-sm">
                {entry!.message}
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
