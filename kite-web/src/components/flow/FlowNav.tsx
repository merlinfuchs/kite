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
import { useCallback, useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";

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

  const [holding, setHolding] = useState(false);
  const [holdProgress, setHoldProgress] = useState(0);
  const [secondsRemaining, setSecondsRemaining] = useState(1.5);
  const [showTapHint, setShowTapHint] = useState(false);
  const holdTimerRef = useRef<NodeJS.Timeout | null>(null);
  const holdStartTimeRef = useRef<number | null>(null);
  const tapHintTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const confirmAndExit = useCallback(() => {
    onExit();
  }, [onExit]);

  const clearHold = useCallback(() => {
    if (holdTimerRef.current) {
      clearInterval(holdTimerRef.current);
      holdTimerRef.current = null;
    }
    holdStartTimeRef.current = null;
    setHolding(false);
    setHoldProgress(0);
    setSecondsRemaining(1.5);
  }, []);

  const handlePointerDown = useCallback(
    (e: React.PointerEvent) => {
      if (e.button !== 0 && e.pointerType === "mouse") return;
      if (tapHintTimeoutRef.current) {
        clearTimeout(tapHintTimeoutRef.current);
        tapHintTimeoutRef.current = null;
      }
      setShowTapHint(false);
      setHolding(true);
      setHoldProgress(0);
      setSecondsRemaining(1.5);
      holdStartTimeRef.current = Date.now();

      holdTimerRef.current = setInterval(() => {
        if (!holdStartTimeRef.current) return;
        const elapsed = Date.now() - holdStartTimeRef.current;
        const remaining = Math.max(0, 1.5 - elapsed / 1000);
        const progress = Math.min(100, (elapsed / 1500) * 100);

        setSecondsRemaining(remaining);
        setHoldProgress(progress);

        if (elapsed >= 1500) {
          clearHold();
          confirmAndExit();
        }
      }, 50);
    },
    [clearHold, confirmAndExit]
  );

  const handlePointerUp = useCallback(() => {
    if (holdStartTimeRef.current) {
      const elapsed = Date.now() - holdStartTimeRef.current;
      clearHold();
      if (elapsed < 1500) {
        setShowTapHint(true);
        if (tapHintTimeoutRef.current) {
          clearTimeout(tapHintTimeoutRef.current);
        }
        tapHintTimeoutRef.current = setTimeout(() => {
          setShowTapHint(false);
        }, 2200);
      }
    } else {
      clearHold();
    }
  }, [clearHold]);

  useEffect(() => {
    return () => {
      if (holdTimerRef.current) clearInterval(holdTimerRef.current);
      if (tapHintTimeoutRef.current) clearTimeout(tapHintTimeoutRef.current);
    };
  }, []);

  return (
    <div className="h-12 flex items-center justify-between px-3 md:px-4 select-none bg-muted/70">
      <div className="flex items-center space-x-2 sm:space-x-4 md:space-x-8">
        <div className="relative">
          <button
            type="button"
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center p-1.5 md:p-0 rounded-md hover:bg-muted md:hover:bg-transparent touch-none select-none"
            onPointerDown={handlePointerDown}
            onPointerUp={handlePointerUp}
            onPointerLeave={clearHold}
            onPointerCancel={clearHold}
            title="Press and hold 1.5s to exit"
            aria-label="Back to App (press and hold 1.5s)"
          >
            <ArrowLeftIcon className="h-5 w-5 flex-none" />
            <span className="hidden md:inline">Back to App</span>
          </button>

          {holding && (
            <div className="absolute top-10 left-0 z-50 min-w-[130px] bg-popover text-popover-foreground border border-border shadow-lg rounded-md px-2.5 py-1.5 text-xs pointer-events-none">
              <div className="flex items-center justify-between font-medium text-[11px] mb-1">
                <span>Hold to exit</span>
                <span>{secondsRemaining.toFixed(1)}s</span>
              </div>
              <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary transition-all duration-75"
                  style={{ width: `${holdProgress}%` }}
                />
              </div>
            </div>
          )}

          {showTapHint && !holding && (
            <div className="absolute top-10 left-0 z-50 whitespace-nowrap bg-popover text-popover-foreground border border-border shadow-lg rounded-md px-2.5 py-1.5 text-xs pointer-events-none">
              Press and hold 1.5s to exit
            </div>
          )}
        </div>
        {isSaving ? (
          <div
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center p-1.5 md:p-0"
            title="Saving Changes"
          >
            <RefreshCwIcon className="h-5 w-5 animate-spin flex-none" />
            <span className="hidden md:inline">Saving Changes</span>
          </div>
        ) : hasUnsavedChanges ? (
          <button
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center p-1.5 md:p-0 rounded-md hover:bg-muted md:hover:bg-transparent"
            onClick={save}
            title="Save Changes"
            aria-label="Save Changes"
          >
            <div className="h-3 w-3 rounded-full bg-foreground/80 flex-none"></div>
            <span className="hidden md:inline">Save Changes</span>
          </button>
        ) : (
          <div
            className="flex space-x-2 text-foreground/70 items-center p-1.5 md:p-0"
            title="No Unsaved Changes"
          >
            <CheckIcon className="h-5 w-5 flex-none" />
            <span className="hidden md:inline">No Unsaved Changes</span>
          </div>
        )}
        {hasUndeployedChanges ? (
          <button
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center disabled:opacity-50 disabled:cursor-not-allowed p-1.5 md:p-0 rounded-md hover:bg-muted md:hover:bg-transparent"
            disabled={hasUnsavedChanges}
            onClick={onDeploy}
            title="Deploy Changes"
            aria-label="Deploy Changes"
          >
            <ArrowUpIcon className="h-5 w-5 flex-none" />
            <span className="hidden md:inline">Deploy Changes</span>
          </button>
        ) : hasUndeployedChanges === false ? (
          <div
            className="flex space-x-2 text-foreground/70 items-center p-1.5 md:p-0"
            title="Changes Deployed"
          >
            <CheckIcon className="h-5 w-5 flex-none" />
            <span className="hidden md:inline">Changes Deployed</span>
          </div>
        ) : null}
      </div>
      <div className="flex items-center space-x-2 sm:space-x-4 md:space-x-6">
        <button
          className={cn(
            "flex space-x-2 items-center p-1.5 md:p-0 rounded-md hover:bg-muted md:hover:bg-transparent",
            chatOpen
              ? "text-foreground"
              : "text-foreground/80 hover:text-foreground"
          )}
          onClick={() => onChatOpenChange(!chatOpen)}
          title="Ask AI"
          aria-label="Ask AI"
        >
          <SparklesIcon className="h-5 w-5 flex-none" />
          <span className="hidden md:inline">Ask AI</span>
        </button>
        {theme === "dark" ? (
          <button
            className="p-1.5 md:p-0 text-foreground/80 hover:text-foreground rounded-md"
            onClick={() => setTheme("light")}
            title="Switch to light theme"
            aria-label="Switch to light theme"
          >
            <MoonStarIcon className="w-5 h-5 md:w-6 md:h-6 cursor-pointer" />
          </button>
        ) : (
          <button
            className="p-1.5 md:p-0 text-foreground/80 hover:text-foreground rounded-md"
            onClick={() => setTheme("dark")}
            title="Switch to dark theme"
            aria-label="Switch to dark theme"
          >
            <SunIcon className="w-5 h-5 md:w-6 md:h-6 cursor-pointer" />
          </button>
        )}
      </div>
    </div>
  );
}
