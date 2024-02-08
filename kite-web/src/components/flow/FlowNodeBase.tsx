import clsx from "clsx";
import { Handle, NodeProps, Position, useEdges } from "reactflow";
import { NodeData } from "./types";
import { LinkIcon } from "@heroicons/react/16/solid";
import { FC, ReactNode, useMemo } from "react";

interface Props extends NodeProps<NodeData> {
  title: string;
  description: string;
  color?: string;
  icon: FC<any>;
  children: ReactNode;
  highlight: boolean;
}

export default function FlowNodeBase(props: Props) {
  const color = props.color ?? "#eab308";

  return (
    <div
      className="pl-2.5 pr-4 py-2.5 shadow-md rounded bg-dark-3 border-dark-3 border-2 cursor-pointer relative"
      style={{
        borderColor: props.selected
          ? "#5457f0"
          : props.highlight
          ? color
          : undefined,
      }}
    >
      <div className="flex items-start space-x-3">
        <div
          className="rounded-md w-8 h-8 flex justify-center items-center"
          style={{ backgroundColor: color }}
        >
          <props.icon className="h-5 w-5 text-white" />
        </div>
        <div>
          <div className="text-sm font-medium text-gray-100 leading-5 mb-1">
            {props.title}
          </div>
          <div className="text-xs text-gray-300">{props.description}</div>
        </div>
      </div>

      {props.children}
    </div>
  );
}
