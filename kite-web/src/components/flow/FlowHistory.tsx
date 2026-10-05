import { FlowVersion } from "@/lib/types/wire.gen";
import { formatDistanceToNow } from "date-fns";
import { HistoryIcon, RefreshCwIcon, RotateCcwIcon } from "lucide-react";
import { useState } from "react";
import { Button } from "../ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";
import { cn } from "@/lib/utils";

interface Props {
  versions?: FlowVersion[];
  hasUnsavedChanges: boolean;
  isLoading: boolean;
  // The version being restored, if any.
  restoringId: string | null;
  canUndoSave: boolean;
  onUndoSave: () => void;
  onRestore: (version: FlowVersion) => void;
}

export default function FlowHistory({
  versions,
  hasUnsavedChanges,
  isLoading,
  restoringId,
  canUndoSave,
  onUndoSave,
  onRestore,
}: Props) {
  const [open, setOpen] = useState(false);
  const busy = restoringId !== null;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            "flex space-x-2 items-center",
            open
              ? "text-foreground"
              : "text-foreground/80 hover:text-foreground"
          )}
        >
          <HistoryIcon className="h-5 w-5" />
          <div>History</div>
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
            disabled={!canUndoSave || busy}
            onClick={onUndoSave}
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
            versions.map((version, i) => (
              <div
                key={version.id}
                className="group flex items-center justify-between gap-3 px-4 py-2 hover:bg-muted/50"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2 text-sm">
                    <span title={new Date(version.created_at).toLocaleString()}>
                      {formatDistanceToNow(new Date(version.created_at), {
                        addSuffix: true,
                      })}
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
                    disabled={busy}
                    onClick={() => onRestore(version)}
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
        </div>
      </PopoverContent>
    </Popover>
  );
}

function versionLabel(version: FlowVersion) {
  if (!version.creator_user_id) return "Before history was kept";

  const kind = version.auto_saved ? "Auto-saved" : "Saved";
  return version.creator_display_name
    ? `${kind} by ${version.creator_display_name}`
    : kind;
}
