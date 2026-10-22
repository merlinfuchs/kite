import { FlowData, NodeProps } from "@/lib/flow/dataSchema";
import { useHookedTheme } from "@/lib/hooks/theme";
import { Node, useReactFlow } from "@xyflow/react";
import {
  AlertCircleIcon,
  ArrowLeftIcon,
  CheckIcon,
  CircleDotIcon,
  CircleOffIcon,
  MoonStarIcon,
  RefreshCwIcon,
  SparklesIcon,
  SunIcon,
} from "lucide-react";
import { useCallback, useEffect } from "react";
import { cn } from "@/lib/utils";
import {
  CommandDeployState,
  commandDeployState,
} from "../app/CommandDeployStatus";
import { useApp } from "@/lib/hooks/api";
import { Command } from "@/lib/types/wire.gen";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

interface Props {
  hasUnsavedChanges: boolean;
  // Commands show whether their changes are in Discord yet.
  command?: Command;
  isSaving: boolean;
  onSave: (d: FlowData) => void;
  onExit: () => void;
  chatOpen: boolean;
  onChatOpenChange: (open: boolean) => void;
}

export default function FlowNav({
  hasUnsavedChanges,
  command,
  isSaving,
  onSave,
  onExit,
  chatOpen,
  onChatOpenChange,
}: Props) {
  const { theme, setTheme } = useHookedTheme();
  const app = useApp();

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
    <div className="h-12 flex items-center justify-between px-3 md:px-4 select-none bg-muted/70">
      <div className="flex items-center space-x-2 sm:space-x-4 md:space-x-8">
        <button
          type="button"
          className="flex space-x-2 text-foreground/80 hover:text-foreground items-center p-1.5 md:p-0 rounded-md hover:bg-muted md:hover:bg-transparent cursor-pointer"
          onClick={onExit}
          title="Back to App"
          aria-label="Back to App"
        >
          <ArrowLeftIcon className="h-5 w-5 flex-none" />
          <span className="hidden md:inline">Back to App</span>
        </button>
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
        {command && app && (
          <DeployStatus
            state={commandDeployState(app, command)}
            error={app.commands_deploy_error?.message}
          />
        )}
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

const deployStatuses = {
  pending: {
    icon: <RefreshCwIcon className="h-5 w-5 animate-spin flex-none" />,
    label: "Updating Discord",
    description: "Your saved changes are on their way to Discord.",
  },
  stopped: {
    icon: <CircleDotIcon className="h-5 w-5 flex-none" />,
    label: "App Stopped",
    description: "Your saved changes go to Discord once your app runs again.",
  },
  error: {
    icon: <AlertCircleIcon className="h-5 w-5 flex-none text-destructive" />,
    label: "Not Updated in Discord",
    description: null,
  },
  disabled: {
    icon: <CircleOffIcon className="h-5 w-5 flex-none" />,
    label: "Disabled",
    description:
      "Disabled commands aren't in Discord. Enable it in the list of commands when it's ready.",
  },
  deployed: {
    icon: <CheckIcon className="h-5 w-5 flex-none" />,
    label: "Live in Discord",
    description:
      "Your saved changes are in Discord. Restart Discord if a change doesn't show up.",
  },
};

function DeployStatus({
  state,
  error,
}: {
  state: CommandDeployState;
  error?: string | null;
}) {
  const status = deployStatuses[state];

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <div className="flex space-x-2 text-foreground/70 items-center p-1.5 md:p-0">
          {status.icon}
          <span className="hidden md:inline">{status.label}</span>
        </div>
      </TooltipTrigger>
      <TooltipContent className="max-w-sm">
        <div className="text-foreground/90">
          {state === "error" ? error : status.description}
        </div>
      </TooltipContent>
    </Tooltip>
  );
}
