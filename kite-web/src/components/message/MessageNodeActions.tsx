import {
  ChevronDownIcon,
  ChevronUpIcon,
  CopyIcon,
  TrashIcon,
} from "lucide-react";
import { useNodeActions } from "@/lib/message/state";
import { Button } from "../ui/button";

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

  const buttons = [
    {
      onClick: actions.moveUp,
      label: "Move up",
      Icon: ChevronUpIcon,
      iconClass: chevron,
    },
    {
      onClick: actions.moveDown,
      label: "Move down",
      Icon: ChevronDownIcon,
      iconClass: chevron,
    },
    {
      onClick: actions.duplicate,
      label: "Duplicate",
      Icon: CopyIcon,
      iconClass: icon,
    },
    {
      onClick: actions.remove,
      label: "Remove",
      Icon: TrashIcon,
      iconClass: icon,
    },
  ];

  return (
    <>
      {buttons.map(
        ({ onClick, label, Icon, iconClass }) =>
          onClick && (
            <Button
              key={label}
              type="button"
              variant="ghost"
              size="icon"
              className={size === "lg" ? "h-9 w-9" : "h-8 w-8"}
              onClick={onClick}
              aria-label={label}
            >
              <Icon className={iconClass} />
            </Button>
          )
      )}
    </>
  );
}
