import { FlowData, NodeData } from "@/lib/flow/dataSchema";
import { useGetFlowData } from "@/lib/flow/parametersStore";
import { useHookedTheme } from "@/lib/hooks/theme";
import { Edge, Node, useStore } from "@xyflow/react";
import {
  AlertCircleIcon,
  ArrowLeftIcon,
  CheckIcon,
  CircleDotIcon,
  CircleOffIcon,
  CoinsIcon,
  MoonStarIcon,
  RefreshCwIcon,
  ShieldAlertIcon,
  SparklesIcon,
  SunIcon,
  TriangleAlertIcon,
} from "lucide-react";
import { ReactNode, useCallback, useEffect, useMemo } from "react";
import { cn } from "@/lib/utils";
import {
  CommandDeployState,
  commandDeployState,
} from "../app/CommandDeployStatus";
import { useApp, useAppStateGuilds } from "@/lib/hooks/api";
import { Command } from "@/lib/types/wire.gen";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import {
  chargedFrequentEvent,
  formatFlowCredits,
  getFlowCredits,
} from "@/lib/flow/credits";
import {
  flowPermissions,
  formatMissingPermissions,
  missingGuildPermissions,
} from "@/lib/flow/templatePermissions";
import { capitalize } from "@/lib/flow/eventTypes";
import env from "@/lib/env/client";
import { BundleInstallNavItem } from "../app/BundleInstallBadge";

interface Props {
  hasUnsavedChanges: boolean;
  // Commands show whether their changes are in Discord yet.
  command?: Command;
  bundleInstallId?: string | null;
  isSaving: boolean;
  onSave: (d: FlowData) => void;
  onExit: () => void;
  chatOpen: boolean;
  onChatOpenChange: (open: boolean) => void;
  // The save history button, if the flow has one.
  history?: ReactNode;
}

export default function FlowNav({
  hasUnsavedChanges,
  command,
  bundleInstallId,
  isSaving,
  onSave,
  onExit,
  chatOpen,
  onChatOpenChange,
  history,
}: Props) {
  const { theme, setTheme } = useHookedTheme();
  const app = useApp();

  const getFlowData = useGetFlowData();

  const save = useCallback(() => {
    if (!isSaving) onSave(getFlowData());
  }, [onSave, getFlowData, isSaving]);

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
          title="Back to app"
          aria-label="Back to app"
        >
          <ArrowLeftIcon className="h-5 w-5 flex-none" />
          <span className="hidden md:inline">Back to app</span>
        </button>
        {isSaving ? (
          <div
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center p-1.5 md:p-0"
            title="Saving changes"
          >
            <RefreshCwIcon className="h-5 w-5 animate-spin flex-none" />
            <span className="hidden md:inline">Saving changes</span>
          </div>
        ) : hasUnsavedChanges ? (
          <button
            className="flex space-x-2 text-foreground/80 hover:text-foreground items-center p-1.5 md:p-0 rounded-md hover:bg-muted md:hover:bg-transparent"
            onClick={save}
            title="Save changes"
            aria-label="Save changes"
          >
            <div className="h-3 w-3 rounded-full bg-foreground/80 flex-none"></div>
            <span className="hidden md:inline">Save changes</span>
          </button>
        ) : (
          <div
            className="flex space-x-2 text-foreground/70 items-center p-1.5 md:p-0"
            title="No unsaved changes"
          >
            <CheckIcon className="h-5 w-5 flex-none" />
            <span className="hidden md:inline">No unsaved changes</span>
          </div>
        )}
        {command && app && (
          <DeployStatus
            state={commandDeployState(app, command)}
            error={app.commands_deploy_error?.message}
          />
        )}
        <BundleInstallNavItem
          bundleInstallId={bundleInstallId}
          className="p-1.5 md:p-0"
        />
      </div>
      <div className="flex items-center space-x-2 sm:space-x-4 md:space-x-6">
        {history}
        <PermissionsStatus />
        <CreditsStatus />
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
    label: "App stopped",
    description: "Your saved changes go to Discord once your app runs again.",
  },
  error: {
    icon: <AlertCircleIcon className="h-5 w-5 flex-none text-destructive" />,
    label: "Not updated in Discord",
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

// Moving or selecting blocks doesn't change what a run costs, so the estimate
// only updates when blocks, their settings or their connections change.
const sameNodes = (a: Node<NodeData>[], b: Node<NodeData>[]) =>
  a.length === b.length &&
  a.every(
    (n, i) => n.id === b[i].id && n.type === b[i].type && n.data === b[i].data
  );
const sameEdges = (a: Edge[], b: Edge[]) =>
  a.length === b.length &&
  a.every(
    (e, i) =>
      e.source === b[i].source &&
      e.target === b[i].target &&
      e.sourceHandle === b[i].sourceHandle
  );

// CreditsStatus shows the most credits a run of the flow costs, and warns
// when a listener for a frequent event pays for every one of them.
function CreditsStatus() {
  const nodes = useStore((s) => s.nodes as Node<NodeData>[], sameNodes);
  const edges = useStore((s) => s.edges, sameEdges);

  const credits = useMemo(() => getFlowCredits(nodes, edges), [nodes, edges]);
  const eventName = useMemo(
    () => chargedFrequentEvent(nodes, edges, credits),
    [nodes, edges, credits]
  );
  const label = formatFlowCredits(credits);

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <a
          href={`${env.NEXT_PUBLIC_DOCS_LINK}/reference/credit-system#estimating-usage`}
          target="_blank"
          className={cn(
            "flex space-x-2 items-center p-1.5 md:p-0",
            eventName
              ? "text-yellow-600 dark:text-yellow-500"
              : "text-foreground/70 hover:text-foreground"
          )}
        >
          {eventName ? (
            <TriangleAlertIcon className="h-5 w-5 flex-none" />
          ) : (
            <CoinsIcon className="h-5 w-5 flex-none" />
          )}
          <span className="md:hidden text-sm">{credits.max}</span>
          <span className="hidden md:inline">
            {credits.max === 0
              ? "Free to run"
              : `${capitalize(label)} ${
                  eventName ? `on every ${eventName}` : "per run"
                }`}
          </span>
        </a>
      </TooltipTrigger>
      <TooltipContent className="max-w-sm">
        <div className="text-foreground/90 space-y-2">
          {eventName && (
            <p>
              This runs for every {eventName} in your servers and costs at least{" "}
              {credits.min} credit{credits.min === 1 ? "" : "s"} each time, even
              when there&apos;s nothing to do. Add an event filter, or a
              condition before the first block that costs credits.
            </p>
          )}
          <p>
            One run of this flow costs {label}. Reading data, conditions and
            loops are free.
            {credits.variableLoops &&
              " Loops are counted once, since how often they run is only known when the flow runs."}{" "}
            Click to learn more.
          </p>
        </div>
      </TooltipContent>
    </Tooltip>
  );
}

// PermissionsStatus warns when the flow uses blocks that need Discord
// permissions the bot doesn't have in its servers.
function PermissionsStatus() {
  const nodes = useStore((s) => s.nodes as Node<NodeData>[], sameNodes);
  const guilds = useAppStateGuilds();

  const missing = useMemo(() => {
    const required = flowPermissions({ nodes, edges: [] });
    return missingGuildPermissions(required, guilds);
  }, [nodes, guilds]);

  if (missing.permissions.length === 0) return null;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <div className="flex space-x-2 items-center p-1.5 md:p-0 text-yellow-600 dark:text-yellow-500">
          <ShieldAlertIcon className="h-5 w-5 flex-none" />
          <span className="hidden md:inline">Missing permissions</span>
        </div>
      </TooltipTrigger>
      <TooltipContent className="max-w-sm">
        <div className="text-foreground/90">
          {formatMissingPermissions(missing)}
        </div>
      </TooltipContent>
    </Tooltip>
  );
}
