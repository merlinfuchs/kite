import { CollapsibleTrigger } from "@radix-ui/react-collapsible";
import { Collapsible, CollapsibleContent } from "@/components/ui/collapsible";
import { ReactNode, useState } from "react";
import { ChevronRightIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import MessageValidationErrorIndicator from "./MessageValidationErrorMarker";
import { useAutoAnimate } from "@formkit/auto-animate/react";

export default function MessageCollapsibleSection({
  children,
  title,
  defaultOpen = true,
  size = "xl",
  valiationPathPrefix,
  actions,
  className,
  animate = true,
}: {
  children: ReactNode;
  title: string;
  defaultOpen?: boolean;
  size?: "sm" | "md" | "lg" | "xl";
  valiationPathPrefix?: string | string[];
  actions?: ReactNode;
  className?: string;
  animate?: boolean;
}) {
  const [open, setOpen] = useState(defaultOpen);

  const textSize = {
    sm: "text-sm",
    md: "text-base",
    lg: "text-lg",
    xl: "text-xl",
  }[size];

  const iconSize = {
    sm: "h-4 w-4",
    md: "h-5 w-5",
    lg: "h-6 w-6",
    xl: "h-7 w-7",
  }[size];

  const [parent] = useAutoAnimate();

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <div className="flex items-center">
        <CollapsibleTrigger
          className={cn(
            "flex items-center space-x-1 flex-auto text-foreground/90",
            textSize
          )}
        >
          <ChevronRightIcon
            className={cn(
              "transition-transform text-muted-foreground",
              iconSize,
              open && "rotate-90"
            )}
          />
          <div>{title}</div>
          {valiationPathPrefix && (
            <div className="pl-1">
              <MessageValidationErrorIndicator
                pathPrefix={valiationPathPrefix}
              />
            </div>
          )}
        </CollapsibleTrigger>
        {actions && (
          <div className="flex-none flex items-center space-x-3">{actions}</div>
        )}
      </div>
      <CollapsibleContent
        className={cn("mt-3", className)}
        ref={animate ? parent : null}
      >
        {children}
      </CollapsibleContent>
    </Collapsible>
  );
}
