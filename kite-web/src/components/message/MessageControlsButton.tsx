import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipPortal,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { forwardRef, ReactNode } from "react";

interface Props {
  icon: any;
  label: string;
  onClick: () => void;
  disabled?: boolean;
  ref: ReactNode;
}

const MessageControlsButton = forwardRef<HTMLButtonElement, Props>(
  (props, ref) => {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="outline"
            size="icon"
            className="rounded-full"
            disabled={props.disabled}
            onClick={props.onClick}
            aria-label={props.label}
            ref={ref}
          >
            <props.icon className="w-5 h-5" />
          </Button>
        </TooltipTrigger>
        {/* Portaled so it isn't clipped by a scroll area it sits at the top
            of, like in the modal editor */}
        <TooltipPortal>
          <TooltipContent>{props.label}</TooltipContent>
        </TooltipPortal>
      </Tooltip>
    );
  }
);

MessageControlsButton.displayName = "MessageControlsButton";
export default MessageControlsButton;
