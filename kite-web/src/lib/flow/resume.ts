import { Edge } from "@xyflow/react";

// The format has to match the backend for the resume point to work.
const componentHandlePrefix = "component_";

export function componentHandleId(componentId: number | undefined) {
  return `${componentHandlePrefix}${componentId}`;
}

// isResumeEdge reports whether the flow suspends between the edge's source and
// target, until a component is used or a modal is submitted.
export function isResumeEdge(edge: Edge, sourceType: string | undefined) {
  return (
    edge.sourceHandle?.startsWith(componentHandlePrefix) ||
    sourceType === "suspend_response_modal"
  );
}
