import { ItemVersion } from "@/lib/types/wire.gen";
import { VersionHistory as History } from "@/lib/hooks/versionHistory";
import { differenceInCalendarDays, format, formatRelative } from "date-fns";
import { HistoryIcon, RefreshCwIcon, RotateCcwIcon } from "lucide-react";
import { Button } from "../ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";
import { cn } from "@/lib/utils";
import { useState } from "react";

// The newest saves are shown until the user asks for all of them.
const collapsedCount = 5;

interface Props {
  history: History;
  isSaving: boolean;
}

export default function VersionHistory({ history, isSaving }: Props) {
  const {
    open,
    setOpen,
    versions,
    isLoading,
    hasUnsavedChanges,
    restoringId,
    canUndoSave,
    undoSave,
    restore,
  } = history;
  const busy = restoringId !== null;
  const [showAll, setShowAll] = useState(false);
  const shown = showAll ? versions : versions?.slice(0, collapsedCount);

  return (
    <Popover
      open={open}
      onOpenChange={(open) => {
        setOpen(open);
        if (!open) setShowAll(false);
      }}
    >
      <PopoverTrigger asChild>
        <button
          className={cn(
            "flex space-x-2 items-center p-1.5 md:p-0 rounded-md hover:bg-muted md:hover:bg-transparent",
            open
              ? "text-foreground"
              : "text-foreground/80 hover:text-foreground"
          )}
          title="History"
          aria-label="History"
        >
          <HistoryIcon className="h-5 w-5 flex-none" />
          <span className="hidden lg:inline">History</span>
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80 p-0">
        <div className="flex items-center justify-between gap-3 border-b px-4 py-3">
          <div>
            <div className="text-sm font-medium">Save history</div>
            <div className="text-xs text-muted-foreground">
              Restoring a save is saved as a new version.
            </div>
          </div>
          <Button
            size="sm"
            variant="secondary"
            className="shrink-0 gap-1.5"
            disabled={!canUndoSave || isSaving || busy}
            onClick={undoSave}
          >
            <RotateCcwIcon className="size-3.5" />
            Undo save
          </Button>
        </div>
        <div className="max-h-80 overflow-y-auto py-1">
          {isLoading ? (
            <div className="px-4 py-3 text-sm text-muted-foreground">
              Loading history...
            </div>
          ) : !versions?.length ? (
            <div className="px-4 py-3 text-sm text-muted-foreground">
              No saves yet. Every save from now on shows up here.
            </div>
          ) : (
            shown?.map((version, i) => (
              <div
                key={version.id}
                className="group flex items-center justify-between gap-3 px-4 py-2 hover:bg-muted/50"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2 text-sm">
                    <span>
                      {formatSaveTime(version.updated_at ?? version.created_at)}
                    </span>
                    {i === 0 && (
                      <span className="rounded bg-primary/15 px-1.5 py-0.5 text-[10px] font-medium uppercase text-primary">
                        {hasUnsavedChanges ? "Last save" : "Current"}
                      </span>
                    )}
                  </div>
                  <div className="truncate text-xs text-muted-foreground">
                    {versionLabel(version)}
                  </div>
                </div>
                {(i !== 0 || hasUnsavedChanges) && (
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-7 shrink-0 px-2 text-xs"
                    disabled={isSaving || busy}
                    onClick={() => restore(version)}
                  >
                    {restoringId === version.id ? (
                      <RefreshCwIcon className="size-3.5 animate-spin" />
                    ) : (
                      "Restore"
                    )}
                  </Button>
                )}
              </div>
            ))
          )}
          {versions && versions.length > collapsedCount && !showAll && (
            <button
              className="w-full px-4 py-2 text-left text-xs text-muted-foreground hover:text-foreground"
              onClick={() => setShowAll(true)}
            >
              Show all {versions.length} saves
            </button>
          )}
          {showAll && (
            <div className="px-4 py-2 text-xs text-muted-foreground">
              Only the last 25 saves are kept.
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}

const sourceLabels: Record<string, string> = {
  editor: "Saved",
  api_token: "Changed with an API token",
  bundle: "Updated from template",
  delete: "Deleted",
};

function versionLabel(version: ItemVersion) {
  if (version.source === "initial") return "Before history was kept";

  const kind = sourceLabels[version.source] ?? "Saved";
  return version.creator_display_name
    ? `${kind} by ${version.creator_display_name}`
    : kind;
}

// Like "Today at 7:42 PM" or "Oct 20 at 7:42 PM", so saves made close
// together can be told apart.
function formatSaveTime(createdAt: string) {
  const date = new Date(createdAt);
  const now = new Date();
  // formatRelative leaves out the time from a week ago on.
  if (differenceInCalendarDays(now, date) >= 6) {
    return format(date, "MMM d 'at' p");
  }

  const text = formatRelative(date, now);
  return text.charAt(0).toUpperCase() + text.slice(1);
}
