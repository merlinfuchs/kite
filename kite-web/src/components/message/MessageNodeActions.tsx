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

  return (
    <>
      {actions.moveUp && (
        <ChevronUpIcon
          className={chevron}
          onClick={actions.moveUp}
          role="button"
        />
      )}
      {actions.moveDown && (
        <ChevronDownIcon
          className={chevron}
          onClick={actions.moveDown}
          role="button"
        />
      )}
      {actions.duplicate && (
        <CopyIcon
          className={cn(icon)}
          onClick={actions.duplicate}
          role="button"
        />
      )}
      <TrashIcon className={icon} onClick={actions.remove} role="button" />
    </>
  );
}
