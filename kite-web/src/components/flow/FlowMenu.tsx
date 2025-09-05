import { LogEntry } from "@/lib/types/wire.gen";
import { cn } from "@/lib/utils";
import {
  BoxIcon,
  GitCompareIcon,
  LucideIcon,
  MessageSquareWarningIcon,
  TextCursorInputIcon,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import FlowLogList from "./FlowLogList";
import FlowNodeEditor from "./FlowNodeEditor";
import FlowNodeExplorer from "./FlowNodeExplorer";
import { useStoreApi } from "@xyflow/react";

type Tab = "action" | "control_flow" | "option" | "logs";

export default function FlowMenu({
  selectedNodeId,
  logs,
}: {
  selectedNodeId: string | null;
  logs?: LogEntry[];
}) {
  const [tab, setTab] = useState<Tab>("action");

  const store = useStoreApi();

  const wrappedSetTab = useCallback(
    (tab: Tab) => {
      setTab(tab);
      store.getState().addSelectedNodes([]);
    },
    [store, setTab]
  );

  return (
    <div className="flex flex-none">
      <div className="flex-none flex flex-col justify-between bg-muted/50">
        <div className="flex-none flex flex-col items-center gap-1">
          <Tab
            id="action"
            icon={BoxIcon}
            title="Action Blocks"
            tab={tab}
            setTab={wrappedSetTab}
          />
          <Tab
            id="control_flow"
            icon={GitCompareIcon}
            title="Control Flow Blocks"
            tab={tab}
            setTab={wrappedSetTab}
          />
          <Tab
            id="option"
            icon={TextCursorInputIcon}
            title="Option Blocks"
            tab={tab}
            setTab={wrappedSetTab}
          />
        </div>
        <div className="flex-none flex flex-col items-center gap-1">
          <Tab
            id="logs"
            icon={MessageSquareWarningIcon}
            title="Logs"
            tab={tab}
            setTab={wrappedSetTab}
          />
        </div>
      </div>
      <div className="flex-none relative w-96 bg-muted/30">
        {(tab === "action" || tab === "control_flow" || tab === "option") && (
          <FlowNodeExplorer category={tab} />
        )}

        {tab === "logs" && <FlowLogList logs={logs} />}

        {selectedNodeId && <FlowNodeEditor nodeId={selectedNodeId} />}
      </div>
    </div>
  );
}

function Tab({
  id,
  icon: Icon,
  title,
  tab,
  setTab,
}: {
  id: Tab;
  icon: LucideIcon;
  title: string;
  tab: Tab;
  setTab: (tab: Tab) => void;
}) {
  return (
    <Tooltip>
      <TooltipTrigger>
        <div
          className={cn(
            "p-3 cursor-pointer",
            tab === id
              ? "text-foreground bg-background/50 border-l-[3px] border-primary"
              : "text-muted-foreground hover:text-foreground hover:bg-background/20"
          )}
          onClick={() => setTab(id)}
          role="button"
        >
          <Icon className="size-8" />
        </div>
      </TooltipTrigger>
      <TooltipContent side="right">{title}</TooltipContent>
    </Tooltip>
  );
}
