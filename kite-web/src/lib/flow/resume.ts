import { Edge } from "@xyflow/react";
import {
  ButtonStyleLink,
  ComponentData,
  ComponentTypeActionRow,
  ComponentTypeButton,
  ComponentTypeChannelSelect,
  ComponentTypeMentionableSelect,
  ComponentTypeRoleSelect,
  ComponentTypeStringSelect,
  ComponentTypeUserSelect,
} from "../types/message.gen";

// The format has to match the backend for the resume point to work.
const componentHandlePrefix = "component_";

export function componentHandleId(componentId: number | undefined) {
  return `${componentHandlePrefix}${componentId}`;
}

// Each option of a string select gets its own output, like a button.
export function selectOptionHandleId(
  componentId: number | undefined,
  optionId: number | undefined
) {
  return `${componentHandlePrefix}${componentId}_option_${optionId}`;
}

// parseComponentHandle returns the ID of the button or select menu of a
// component output, also of the output of a select menu option.
export function parseComponentHandle(handle: string | null | undefined) {
  const match = handle?.match(/^component_(\d+)(?:_option_\d+)?$/);
  return match ? parseInt(match[1], 10) : undefined;
}

const selectMenuComponentTypes = [
  ComponentTypeStringSelect,
  ComponentTypeUserSelect,
  ComponentTypeRoleSelect,
  ComponentTypeMentionableSelect,
  ComponentTypeChannelSelect,
];

export function isSelectMenu(c: ComponentData) {
  return selectMenuComponentTypes.includes(c.type ?? 0);
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
    isSelectMenu(c) ||
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

// The options of a string select that get their own output.
export function selectMenuOptions(c: ComponentData) {
  if (c.type !== ComponentTypeStringSelect) return [];
  return (c.options || []).filter((o) => o.id !== undefined);
}

// The IDs of the outputs a message block has, one per button, select menu and
// string select option. The select menu's own output is kept for flows that
// branch on interaction.value, it runs when no picked option is connected.
export function getComponentHandleIds(components: ComponentData[]) {
  return collectComponentGroups(components)
    .flat()
    .flatMap((c) => [
      componentHandleId(c.id),
      ...selectMenuOptions(c).map((o) => selectOptionHandleId(c.id, o.id)),
    ]);
}
