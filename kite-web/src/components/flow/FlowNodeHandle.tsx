import { primaryColor } from "@/lib/flow/nodes";
import { Handle, Position } from "reactflow";

interface Props {
  type: "source" | "target";
  position: Position;
  color?: string;
}

export default function FlowNodeHandle({ type, position, color }: Props) {
  return (
    <Handle
      type={type}
      position={position}
      className="rounded-full"
      style={{
        backgroundColor: color ?? primaryColor,
        translate: position === Position.Top ? "0 -2px" : "0 2px",
        height: "10px",
        width: "10px",
      }}
    />
  );
}
