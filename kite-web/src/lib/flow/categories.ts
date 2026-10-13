import { FlowContextType } from "./context";
import { blockDefinitions } from "../blocks";
import { getNodeValues } from "./nodes";

export interface NodeCategorySection {
  title: string;
  nodeTypes: string[];
  contextTypes: FlowContextType[] | null;
  // Flow types the section is shown first in, as their most important blocks.
  pinnedIn?: FlowContextType[];
}

const sections: Record<
  "option" | "action" | "control_flow",
  Omit<NodeCategorySection, "nodeTypes">[]
> = {
  option: [
    {
      title: "Commands",
      contextTypes: ["command"],
    },
    {
      title: "Events",
      contextTypes: ["event_discord"],
    },
    /* {
      title: "Events",
      nodeTypes: ["option_event_filter"],
    }, */
  ],
  action: [
    {
      title: "Responses",
      // Functions can respond to the interaction of the flow that calls them.
      contextTypes: [
        "command",
        "component_button",
        "component_select_menu",
        "function",
      ],
    },
    {
      title: "Messages",
      contextTypes: null,
    },
    {
      title: "Members",
      contextTypes: null,
    },
    {
      title: "Users",
      contextTypes: null,
    },
    {
      title: "Roles",
      contextTypes: null,
    },

    {
      title: "Servers",
      contextTypes: null,
    },
    {
      title: "Channels",
      contextTypes: null,
    },
    {
      title: "Voice",
      contextTypes: null,
    },
    {
      title: "Bot",
      contextTypes: null,
    },
    {
      title: "Functions",
      contextTypes: null,
    },
    {
      title: "Stored Variables",
      contextTypes: null,
    },
    {
      title: "Roblox",
      contextTypes: null,
    },
    {
      title: "AI",
      contextTypes: null,
    },
    {
      title: "API Requests",
      contextTypes: null,
    },
    {
      title: "Utilities",
      contextTypes: null,
    },
  ],
  control_flow: [
    {
      title: "Conditions",
      contextTypes: null,
    },
    {
      title: "Loops",
      contextTypes: null,
    },
    {
      title: "Errors",
      contextTypes: null,
    },
    {
      title: "Others",
      contextTypes: null,
    },
    {
      // Only has the Return block, which only exists in functions. It's what
      // makes a function useful, so it comes first there.
      title: "Functions",
      contextTypes: null,
      pinnedIn: ["function"],
    },
  ],
};

// A misspelled category would hide a block from the explorer, and make it
// available in every kind of flow.
const sectionTitles = new Set(
  Object.values(sections).flatMap((list) => list.map((s) => s.title))
);
for (const block of blockDefinitions) {
  if (block.category && !sectionTitles.has(block.category)) {
    throw new Error(`Unknown category of ${block.type}: ${block.category}`);
  }
}

// The tab a block is listed in when a section of the same title exists in
// several tabs, like Functions: control blocks under Control Flow, the others
// under Actions.
function blockTab(type: string): keyof typeof sections {
  if (type.startsWith("control_")) return "control_flow";
  if (type.startsWith("option_")) return "option";
  return "action";
}

const titleCounts = new Map<string, number>();
for (const list of Object.values(sections)) {
  for (const s of list) {
    titleCounts.set(s.title, (titleCounts.get(s.title) ?? 0) + 1);
  }
}

// The blocks of each section are those whose definition names it, in the
// order of the definitions.
export const nodeCategories = Object.fromEntries(
  Object.entries(sections).map(([category, list]) => [
    category,
    list.map((section) => ({
      ...section,
      nodeTypes: blockDefinitions
        .filter(
          (b) =>
            b.category === section.title &&
            (titleCounts.get(section.title) === 1 ||
              blockTab(b.type) === category)
        )
        .map((b) => b.type),
    })),
  ])
) as Record<keyof typeof sections, NodeCategorySection[]>;

export type NodeCategory = keyof typeof nodeCategories;

function isSectionAvailable(
  section: NodeCategorySection,
  context: FlowContextType
) {
  return !section.contextTypes || section.contextTypes.includes(context);
}

// The sections of a category in the order they're shown in a flow type:
// sections pinned in it first, then the others in their usual order.
export function orderSections<T extends { pinnedIn?: FlowContextType[] }>(
  sections: T[],
  context: FlowContextType
): T[] {
  const pinned = (s: T) => !!s.pinnedIn?.includes(context);
  return [...sections.filter(pinned), ...sections.filter((s) => !pinned(s))];
}

// A block's own contexts take precedence over the sections it's listed in.
// Blocks in neither (e.g. condition items) are always available, as they only
// exist as children of other blocks.
export function isNodeTypeAvailable(type: string, context: FlowContextType) {
  const contexts = getNodeValues(type).contexts;
  if (contexts) return contexts.includes(context);

  const sections = Object.values(nodeCategories)
    .flat()
    .filter((s) => s.nodeTypes.includes(type));
  return (
    sections.length === 0 ||
    sections.some((s) => isSectionAvailable(s, context))
  );
}
