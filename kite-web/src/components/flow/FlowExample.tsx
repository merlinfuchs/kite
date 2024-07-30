import React from "react";
import { ReactFlow } from "@xyflow/react";

import "@xyflow/react/dist/base.css";
import { useHookedTheme } from "@/lib/hooks/theme";
import { edgeTypes, nodeTypes } from "@/lib/flow/components";

const initialNodes = [
  {
    id: "1",
    position: { x: 0, y: 0 },
    data: { name: "ban", description: "Ban a user" },
    type: "entry_command",
  },
  {
    id: "5",
    position: { x: 100, y: 150 },
    data: {
      member_target: "1",
    },
    type: "action_member_ban",
  },
  {
    id: "2",
    position: { x: 0, y: 300 },
    data: {
      message_data: {
        content: "You have been banned from the server.",
      },
    },
    type: "action_response_create",
  },
  {
    id: "3",
    position: { x: -50, y: -150 },
    data: {
      name: "user",
      description: "The user you want to ban",
      command_argument_type: "user",
      command_argument_required: true,
    },
    type: "option_command_argument",
  },
  {
    id: "4",
    position: { x: 250, y: -150 },
    data: {
      name: "reason",
      description: "Why you want to ban the user",
      command_argument_type: "string",
    },
    type: "option_command_argument",
  },
];

const initialEdges = [
  { id: "e1-5", source: "1", target: "5", type: "fixed" },
  { id: "e5-2", source: "5", target: "2", type: "fixed" },
  { id: "e3-1", source: "3", target: "1", type: "fixed" },
  { id: "e4-1", source: "4", target: "1", type: "fixed" },
];

export default function FlowExample() {
  const { theme } = useHookedTheme();

  return (
    <div style={{ width: "100%", height: "100%" }} className="nowheel">
      <ReactFlow
        nodes={initialNodes}
        edges={initialEdges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        elementsSelectable={false}
        nodesConnectable={false}
        nodesDraggable={false}
        connectOnClick={false}
        draggable={false}
        panOnDrag={false}
        zoomOnScroll={false}
        zoomOnPinch={false}
        colorMode={theme === "dark" ? "dark" : "light"}
        className="!bg-transparent"
        fitView
      />
    </div>
  );
}
