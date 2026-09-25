import { Edge } from "@xyflow/react";
import {
  ButtonStyleLink,
  ComponentData,
  ComponentTypeActionRow,
  ComponentTypeButton,
  ComponentTypeStringSelect,
} from "../types/message.gen";

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

// Groups interactive components the way Discord lays them out: one group per action row and one per section accessory.
export function collectComponentGroups(
  components: ComponentData[]
): ComponentData[][] {
  const groups: ComponentData[][] = [];

  const isInteractive = (c: ComponentData) =>
    c.type === ComponentTypeStringSelect ||
    (c.type === ComponentTypeButton && c.style !== ButtonStyleLink);

  const walk = (c: ComponentData) => {
    if (c.type === ComponentTypeActionRow) {
      const interactive = (c.components || []).filter(isInteractive);
      if (interactive.length > 0) groups.push(interactive);
      return;
    }

    c.components?.forEach(walk);
    if (c.accessory && isInteractive(c.accessory)) groups.push([c.accessory]);
  };

  components.forEach(walk);
  return groups;
}

// The IDs of the outputs a message block has, one per button or select menu.
export function getComponentHandleIds(components: ComponentData[]) {
  return collectComponentGroups(components)
    .flat()
    .map((c) => componentHandleId(c.id));
}
