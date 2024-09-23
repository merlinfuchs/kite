import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
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
            ref={ref}
          >
            <props.icon className="w-5 h-5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent>{props.label}</TooltipContent>
      </Tooltip>
    );
  }
);

MessageControlsButton.displayName = "MessageControlsButton";
export default MessageControlsButton;
