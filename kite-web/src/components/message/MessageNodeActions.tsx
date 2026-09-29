import {
  ChevronDownIcon,
  ChevronUpIcon,
  CopyIcon,
  TrashIcon,
} from "lucide-react";
import { useNodeActions } from "@/lib/message/state";
import { cn } from "@/lib/utils";

/** The move, duplicate and remove buttons in the header of a node's section. */
export default function MessageNodeActions({
  actions,
  size = "md",
}: {
  actions: ReturnType<typeof useNodeActions>;
  size?: "md" | "lg";
}) {
  const chevron = size === "lg" ? "h-6 w-6" : "h-5 w-5";
  const icon = size === "lg" ? "h-5 w-5" : "h-4 w-4";
  const button = cn(
    "inline-flex items-center justify-center rounded-md hover:bg-muted",
    size === "lg" ? "h-9 w-9" : "h-8 w-8"
  );

  return (
    <>
      {actions.moveUp && (
        <button
          type="button"
          className={button}
          onClick={actions.moveUp}
          aria-label="Move up"
        >
          <ChevronUpIcon className={chevron} />
        </button>
      )}
      {actions.moveDown && (
        <button
          type="button"
          className={button}
          onClick={actions.moveDown}
          aria-label="Move down"
        >
          <ChevronDownIcon className={chevron} />
        </button>
      )}
      {actions.duplicate && (
        <button
          type="button"
          className={button}
          onClick={actions.duplicate}
          aria-label="Duplicate"
        >
          <CopyIcon className={icon} />
        </button>
      )}
      {actions.remove && (
        <button
          type="button"
          className={button}
          onClick={actions.remove}
          aria-label="Remove"
        >
          <TrashIcon className={icon} />
        </button>
      )}
    </>
  );
}
