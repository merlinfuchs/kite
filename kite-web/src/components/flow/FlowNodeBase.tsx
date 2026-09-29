import { NodeProps } from "@/lib/flow/dataSchema";
import { ReactNode } from "react";
import { errorColor, primaryColor, useNodeValues } from "@/lib/flow/nodes";
import { useMissingIntegrations } from "@/lib/integrations/hooks";
import { useAppId } from "@/lib/hooks/params";
import Link from "next/link";
import FlowNodeMarkers from "./FlowNodeMarkers";
import DynamicIcon from "../icons/DynamicIcon";
import { useChangedByAI } from "@/lib/flow/context";

interface Props extends NodeProps {
  title?: string;
  description?: string;
  children: ReactNode;
  highlight?: boolean;
  showConnectedMarker?: boolean;
  color?: string;
  showId?: boolean;
}

export default function FlowNodeBase(props: Props) {
  const {
    color: defaultColor,
    icon,
    defaultTitle,
    defaultDescription,
  } = useNodeValues(props.type);

  const color = props.color || defaultColor;
  const changedByAI = useChangedByAI(props.id);
  const appId = useAppId();
  const missingIntegrations = useMissingIntegrations(props.type);

  return (
    <div
      className="pl-2.5 pr-4 py-2.5 shadow-md rounded bg-muted border-2 relative max-w-sm min-w-32 cursor-grab group"
      style={{
        borderColor: props.selected
          ? primaryColor
          : missingIntegrations.length > 0
          ? errorColor
          : props.highlight
          ? color
          : undefined,
        boxShadow: changedByAI ? `0 0 0 4px ${primaryColor}66` : undefined,
      }}
    >
      {props.showId && (
        <div className="text-[9px] font-light text-foreground/90 absolute -top-6 right-0 bg-muted rounded-[3px] px-1 py-0.5 max-w-24 truncate hidden group-hover:block">
          {props.id}
        </div>
      )}

      <div className="flex items-start space-x-3">
        <div
          className="rounded-md w-8 h-8 flex justify-center items-center flex-none"
          style={{ backgroundColor: color }}
        >
          <DynamicIcon name={icon as any} className="h-5 w-5 text-white" />
        </div>
        <div className="overflow-hidden">
          <div className="text-sm font-medium text-foreground leading-5 mb-1 truncate">
            {props.title || props.data.custom_label || defaultTitle}
          </div>
          <div className="text-xs text-muted-foreground">
            {props.description || defaultDescription}
          </div>
        </div>
      </div>

      {props.children}

      {missingIntegrations.length > 0 && (
        <div className="text-xs text-red-600 dark:text-red-400 mt-2">
          {missingIntegrations.map((i) => i.name).join(", ")}{" "}
          {missingIntegrations.length === 1 ? "isn't" : "aren't"} connected.{" "}
          <Link
            href={{
              pathname: "/apps/[appId]/integrations",
              query: { appId },
            }}
            target="_blank"
            className="underline"
          >
            Connect {missingIntegrations.length === 1 ? "it" : "them"}
          </Link>
        </div>
      )}

      <FlowNodeMarkers {...props} />
    </div>
  );
}
