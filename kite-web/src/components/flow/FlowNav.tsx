import { FlowData, NodeProps } from "@/lib/flow/dataSchema";
import { useHookedTheme } from "@/lib/hooks/theme";
import { Node, useReactFlow } from "@xyflow/react";
import {
  ArrowLeftIcon,
  ArrowUpIcon,
  CheckIcon,
  MoonStarIcon,
  RefreshCwIcon,
  SparklesIcon,
  SunIcon,
} from "lucide-react";
import { ReactNode, useCallback, useEffect } from "react";
import { cn } from "@/lib/utils";
import { Switch } from "../ui/switch";

interface Props {
  hasUnsavedChanges: boolean;
  hasUndeployedChanges?: boolean;
  isSaving: boolean;
  isDeploying?: boolean;
  onSave: (d: FlowData) => void;
  onDeploy?: () => void;
  onExit: () => void;
  chatOpen: boolean;
  onChatOpenChange: (open: boolean) => void;
  autoSave: boolean;
  onAutoSaveChange: (enabled: boolean) => void;
  // The save history button, if the flow has one.
  history?: ReactNode;
}

export default function FlowNav({
  hasUnsavedChanges,
  hasUndeployedChanges,
  isSaving,
  onSave,
  onDeploy,
  onExit,
  chatOpen,
  onChatOpenChange,
  autoSave,
  onAutoSaveChange,
  history,
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
          <div className="flex space-x-2 text-foreground/70 items-center">
            <CheckIcon className="h-5 w-5" />
            <div>{autoSave ? "All Changes Saved" : "No Unsaved Changes"}</div>
          </div>
        )}
        {hasUndeployedChanges ? (
          <button
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={hasUnsavedChanges}
            onClick={onDeploy}
          >
            <ArrowUpIcon className="h-5 w-5" />
            <div>Deploy Changes</div>
          </button>
        ) : hasUndeployedChanges === false ? (
          <div className="flex space-x-2 text-foreground/70 items-center">
            <CheckIcon className="h-5 w-5" />
            <div>Changes Deployed</div>
          </div>
        ) : null}
      </div>
      <div className="flex items-center space-x-6">
        <label
          className="flex space-x-2 text-foreground/80 hover:text-foreground items-center cursor-pointer"
          title="Save this flow automatically a few seconds after each change"
        >
          <Switch
            checked={autoSave}
            onCheckedChange={onAutoSaveChange}
            className="scale-75"
          />
          <div>Auto-save</div>
        </label>
        {history}
        <button
          className={cn(
            "flex space-x-2 items-center",
            chatOpen
              ? "text-foreground"
              : "text-foreground/80 hover:text-foreground"
          )}
          onClick={() => onChatOpenChange(!chatOpen)}
        >
          <SparklesIcon className="h-5 w-5" />
          <div>Ask AI</div>
        </button>
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
