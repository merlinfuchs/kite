import { FlowContextType } from "./context";
import { blockDefinitions } from "../blocks";
import { getNodeValues } from "./nodes";

export interface NodeCategorySection {
  title: string;
  nodeTypes: string[];
  contextTypes: FlowContextType[] | null;
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
      contextTypes: ["command", "component_button", "component_select_menu"],
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
      title: "Stored Variables",
      contextTypes: null,
    },
    {
      title: "Roblox",
      contextTypes: null,
    },
    {
      title: "Cookie API",
      contextTypes: null,
    },
    {
      title: "ER:LC",
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

// The blocks of each section are those whose definition names it, in the
// order of the definitions.
export const nodeCategories = Object.fromEntries(
  Object.entries(sections).map(([category, list]) => [
    category,
    list.map((section) => ({
      ...section,
      nodeTypes: blockDefinitions
        .filter((b) => b.category === section.title)
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
