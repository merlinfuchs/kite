import { FlowData, NodeProps } from "@/lib/flow/dataSchema";
import { useHookedTheme } from "@/lib/hooks/theme";
import { Node, useReactFlow } from "@xyflow/react";
import {
  ArrowLeftIcon,
  ArrowUpIcon,
  CheckIcon,
  MoonStarIcon,
  RefreshCwIcon,
  SunIcon,
} from "lucide-react";
import { useCallback, useEffect } from "react";

interface Props {
  hasUnsavedChanges: boolean;
  hasUndeployedChanges?: boolean;
  isSaving: boolean;
  isDeploying?: boolean;
  onSave: (d: FlowData) => void;
  onDeploy?: () => void;
  onExit: () => void;
}

export default function FlowNav({
  hasUnsavedChanges,
  hasUndeployedChanges,
  isSaving,
  onSave,
  onDeploy,
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
          <div className="flex space-x-2 text-foreground/70 items-center">
            <CheckIcon className="h-5 w-5" />
            <div>No Unsaved Changes</div>
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
