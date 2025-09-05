import { FlowData, NodeData, NodeProps } from "@/lib/flow/dataSchema";
import { useCallback, useEffect } from "react";
import { Node, useReactFlow } from "@xyflow/react";
import {
  ArrowLeftIcon,
  CheckIcon,
  LogsIcon,
  MessageSquareWarningIcon,
  MoonStarIcon,
  RefreshCwIcon,
  SunIcon,
} from "lucide-react";
import { useHookedTheme } from "@/lib/hooks/theme";

interface Props {
  hasUnsavedChanges: boolean;
  isSaving: boolean;
  onSave: (d: FlowData) => void;
  onExit: () => void;
}

export default function FlowNav({
  hasUnsavedChanges,
  isSaving,
  onSave,
  onExit,
}: Props) {
  const { theme, setTheme } = useHookedTheme();

  const { getEdges, getNodes } = useReactFlow<Node<NodeProps>>();

  const save = useCallback(() => {
    onSave({
      nodes: getNodes(),
      edges: getEdges(),
    });
  }, [onSave, getNodes, getEdges]);

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        save();
      }
    }

    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [onSave, save]);

  return (
    <div className="h-12 flex items-center justify-between px-4 select-none bg-muted/70">
      <div className="flex items-center space-x-8">
        <button
          className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
          onClick={onExit}
        >
          <ArrowLeftIcon className="h-5 w-5" />
          <div>Back to App</div>
        </button>
        {isSaving ? (
          <div
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
            onClick={save}
          >
            <RefreshCwIcon className="h-5 w-5 animate-spin" />
            <div>Saving Changes</div>
          </div>
        ) : hasUnsavedChanges ? (
          <button
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center"
            onClick={save}
          >
            <div className="h-3 w-3 rounded-full bg-foreground/80"></div>
            <div>Save Changes</div>
          </button>
        ) : (
          <div className="flex space-x-2 text-foreground/80 items-center">
            <CheckIcon className="h-5 w-5" />
            <div>No Unsaved Changes</div>
          </div>
        )}
      </div>
      {/*isDeploying ? (
        <div className="flex space-x-2 text-foreground hover:text-white items-center">
          <RefreshCwIcon className="h-5 w-5 animate-spin" />
          <div>Deploying Changes</div>
        </div>
      ) : (
        <button
          className="flex space-x-2 text-foreground hover:text-white items-center"
          onClick={deploy}
        >
          <ArrowUpIcon className="h-5 w-5" />
          <div>Deploy Changes</div>
        </button>
      )*/}
      <div>
        {theme === "dark" ? (
          <MoonStarIcon
            className="w-6 h-6 cursor-pointer"
            onClick={() => setTheme("light")}
          />
        ) : (
          <SunIcon
            className="w-6 h-6 cursor-pointer"
            onClick={() => setTheme("dark")}
          />
        )}
      </div>
    </div>
  );
}
