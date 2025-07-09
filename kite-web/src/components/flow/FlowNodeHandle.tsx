import { Handle, Position } from "@xyflow/react";
import { primaryColor } from "@/lib/flow/nodes";
import { cn } from "@/lib/utils";

interface Props {
  type: "source" | "target";
  position: Position;
  color?: string;
  isConnectable?: boolean;
  size?: "small" | "medium" | "large";
  id?: string;
}

export default function FlowNodeHandle({
  type,
  position,
  color,
  isConnectable,
  size = "medium",
  id,
}: Props) {
  const sizeMap = {
    small: "10px",
    medium: "12px",
    large: "14px",
  };

  return (
    <Handle
      id={id}
      type={type}
      position={position}
      isConnectable={isConnectable}
      className="rounded-full"
      style={{
        backgroundColor: color ?? primaryColor,
        translate:
          position === Position.Top
            ? "0 -3px"
            : position === Position.Bottom
            ? "0 3px"
            : position === Position.Left
            ? "-2px 0"
            : "2px 0",
        height: sizeMap[size],
        width: sizeMap[size],
      }}
    />
  );
}
