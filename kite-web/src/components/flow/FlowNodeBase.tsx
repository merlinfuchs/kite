import { NodeProps } from "@/lib/flow/data";
import { ReactNode } from "react";
import { primaryColor, useNodeValues } from "@/lib/flow/nodes";
import FlowNodeMarkers from "./FlowNodeMarkers";

interface Props extends NodeProps {
  title?: string;
  description?: string;
  children: ReactNode;
  highlight?: boolean;
  showConnectedMarker?: boolean;
  color?: string;
}

export default function FlowNodeBase(props: Props) {
  const {
    color: defaultColor,
    icon: Icon,
    defaultTitle,
    defaultDescription,
  } = useNodeValues(props.type);

  const color = props.color || defaultColor;

  return (
    <div
      className="pl-2.5 pr-4 py-2.5 shadow-md rounded bg-muted border-2 relative max-w-sm min-w-32 cursor-grab"
      style={{
        borderColor: props.selected
          ? primaryColor
          : props.highlight
          ? color
          : undefined,
      }}
    >
      <div className="flex items-start space-x-3">
        <div
          className="rounded-md w-8 h-8 flex justify-center items-center flex-none"
          style={{ backgroundColor: color }}
        >
          <Icon className="h-5 w-5 text-white" />
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

      <FlowNodeMarkers {...props} />
    </div>
  );
}
